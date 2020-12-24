package audit

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	util2 "github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-api-go/pkg/container"
	crypto "github.com/nspcc-dev/neofs-crypto"
	"github.com/nspcc-dev/neofs-node/pkg/innerring/invoke"
	"github.com/nspcc-dev/neofs-node/pkg/innerring/rpc"
	"github.com/nspcc-dev/neofs-node/pkg/morph/client"
	containerWrapper "github.com/nspcc-dev/neofs-node/pkg/morph/client/container/wrapper"
	netmapWrapper "github.com/nspcc-dev/neofs-node/pkg/morph/client/netmap/wrapper"
	"github.com/nspcc-dev/neofs-node/pkg/services/audit"
	"github.com/nspcc-dev/neofs-node/pkg/services/audit/auditor"
	"github.com/nspcc-dev/neofs-node/pkg/util"
	"github.com/nspcc-dev/neofs-node/pkg/util/logger/test"
	"github.com/panjf2000/ants/v2"
	"github.com/stretchr/testify/require"
)

const (
	cid = "FANTWjDrYt5b1jqHThjAW5v33hmRz4xzxJvXajHSEUUX"

	containerContractStr = "f42b41702a19e681b467bba63ce42c8f5aeca679"
	netmapContractStr    = "c47689c6020a7904aa86c3bc85986564ffff9e0a"

	morphEnpoint = "http://morph_chain.neofs.devenv:30333"

	irKeyWIF = "L3o221BojgcCPYgdbXsm6jn7ayTZ72xwREvBHXKknR8VJ3G4WmjB"

	pdpPoolSize = 10
	porPoolSize = 10

	maxPDPSleep = 5 * time.Second

	sgSearchTimeout = 5 * time.Second
)

type taskManager struct {
	TaskManager

	ctxPrm auditor.ContextPrm

	pdpPool, porPool util.WorkerPool
}

func printField(name string, val interface{}) {
	const pad = 20

	if name != "" {
		name += ":"
	}

	fmt.Printf(fmt.Sprintf("%%-%dv%%v\n", pad), name, val)
}

func (m *taskManager) WriteReport(r *audit.Report) error {
	res := r.Result()

	printField("Version", res.Version())
	printField("Epoch", res.AuditEpoch())
	printField("Container", res.ContainerID())
	printField("Complete", res.Complete())
	printField("PoR requests", res.Requests())
	printField("PoR retries", res.Retries())

	fmt.Println("SG passed PoR:")
	if ids := res.PassSG(); len(ids) > 0 {
		for i := range ids {
			printField("", ids[i])
		}
	} else {
		printField("", "<empty>")
	}

	fmt.Println("SG passed PoR:")
	if ids := res.FailSG(); len(ids) > 0 {
		for i := range ids {
			printField("", ids[i])
		}
	} else {
		printField("", "<empty>")
	}

	printField("PoP hit", res.Hit())
	printField("PoP miss", res.Miss())
	printField("PoP fail", res.Fail())

	fmt.Println("Nodes passed PDP:")
	if ids := res.PassNodes(); len(ids) > 0 {
		for i := range ids {
			printField("", hex.EncodeToString(ids[i]))
		}
	} else {
		printField("", "<empty>")
	}

	fmt.Println("Nodes failed PDP:")
	if ids := res.FailNodes(); len(ids) > 0 {
		for i := range ids {
			printField("", hex.EncodeToString(ids[i]))
		}
	} else {
		printField("", "<empty>")
	}

	return nil
}

func (m *taskManager) PushTask(task *audit.Task) error {
	auditor.NewContext(m.ctxPrm).
		WithPDPWorkerPool(m.pdpPool).
		WithPoRWorkerPool(m.porPool).
		WithTask(task).
		Execute()

	return nil
}

func morphClient(t *testing.T, key *ecdsa.PrivateKey) *client.Client {
	c, err := client.New(key, morphEnpoint)
	require.NoError(t, err)

	return c
}

func netmapClient(t *testing.T, c *client.Client) *netmapWrapper.Wrapper {
	contractHash, err := util2.Uint160DecodeStringLE(netmapContractStr)
	require.NoError(t, err)

	cli, err := invoke.NewNoFeeNetmapClient(c, contractHash)
	require.NoError(t, err)

	return cli
}

func containerClient(t *testing.T, c *client.Client) *containerWrapper.Wrapper {
	contractHash, err := util2.Uint160DecodeStringLE(containerContractStr)
	require.NoError(t, err)

	cli, err := invoke.NewNoFeeContainerClient(c, contractHash)
	require.NoError(t, err)

	return cli
}

func TestContainerAudit(t *testing.T) {
	log := test.NewLogger(true)

	irKey, err := crypto.WIFDecode(irKeyWIF)
	require.NoError(t, err)

	cli := rpc.New(
		rpc.WithLogger(log),
		rpc.WithKey(irKey),
	)

	auditCtxPrm := auditor.ContextPrm{}
	auditCtxPrm.SetMaxPDPSleep(maxPDPSleep)
	auditCtxPrm.SetLogger(log)
	auditCtxPrm.SetContainerCommunicator(cli)

	pdpPool, err := ants.NewPool(pdpPoolSize)
	require.NoError(t, err)

	porPool, err := ants.NewPool(porPoolSize)
	require.NoError(t, err)

	manager := &taskManager{
		ctxPrm:  auditCtxPrm,
		pdpPool: pdpPool,
		porPool: porPool,
	}

	cMorph := morphClient(t, irKey)

	nmCli := netmapClient(t, cMorph)

	cnrCli := containerClient(t, cMorph)

	proc := &Processor{
		log:             log,
		sgSearcher:      cli,
		searchTimeout:   sgSearchTimeout,
		containerClient: cnrCli,
		netmapClient:    nmCli,
		taskManager:     manager,
		reporter:        manager,
	}

	containerID := container.NewID()
	require.NoError(t, containerID.Parse(cid))

	nm, err := nmCli.GetNetMap(0)
	require.NoError(t, err)

	epoch, err := nmCli.Epoch()
	require.NoError(t, err)

	task := new(audit.Task).
		WithNetworkMap(nm).
		WithAuditContext(context.Background())

	proc.processContainer(task, epoch, containerID)
}
