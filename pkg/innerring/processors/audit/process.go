package audit

import (
	"context"

	"github.com/nspcc-dev/neofs-api-go/pkg/container"
	"github.com/nspcc-dev/neofs-api-go/pkg/netmap"
	"github.com/nspcc-dev/neofs-api-go/pkg/object"
	"github.com/nspcc-dev/neofs-node/pkg/services/audit"
	"github.com/nspcc-dev/neofs-node/pkg/util/rand"
	"go.uber.org/zap"
)

func (ap *Processor) processStartAudit(epoch uint64) {
	log := ap.log.With(zap.Uint64("epoch", epoch))

	ap.prevAuditCanceler()

	skipped := ap.taskManager.Reset()
	if skipped > 0 {
		ap.log.Info("some tasks from previous epoch are skipped",
			zap.Int("amount", skipped),
		)
	}

	containers, err := ap.selectContainersToAudit(epoch)
	if err != nil {
		log.Error("container selection failure", zap.String("error", err.Error()))

		return
	}

	log.Info("select containers for audit", zap.Int("amount", len(containers)))

	nm, err := ap.netmapClient.GetNetMap(0)
	if err != nil {
		ap.log.Error("can't fetch network map",
			zap.String("error", err.Error()))

		return
	}

	var auditCtx context.Context
	auditCtx, ap.prevAuditCanceler = context.WithCancel(context.Background())

	for i := range containers {
		auditTask := new(audit.Task).
			WithAuditContext(auditCtx).
			WithNetworkMap(nm)

		ap.processContainer(auditTask, epoch, containers[i])
	}
}

func (ap *Processor) processContainer(task *audit.Task, epoch uint64, cid *container.ID) {
	log := ap.log.With(
		zap.Stringer("cid", cid),
	)

	cnr, err := ap.containerClient.Get(cid) // get container structure
	if err != nil {
		log.Error("can't get container info, ignore",
			zap.String("error", err.Error()))

		return
	}

	// find all container nodes for current epoch
	nodes, err := task.NetworkMap().GetContainerNodes(cnr.PlacementPolicy(), nil)
	if err != nil {
		log.Info("can't build placement for container, ignore",
			zap.String("error", err.Error()))

		return
	}

	n := nodes.Flatten()
	crand := rand.New() // math/rand with cryptographic source

	// shuffle nodes to ask a random one
	crand.Shuffle(len(n), func(i, j int) {
		n[i], n[j] = n[j], n[i]
	})

	// search storage groups
	storageGroups := ap.findStorageGroups(cid, n)

	log.Info("select storage groups for audit",
		zap.Int("amount", len(storageGroups)))

	task.
		WithReporter(&epochAuditReporter{
			epoch: epoch,
			rep:   ap.reporter,
		}).
		WithStorageGroupList(storageGroups).
		WithContainerStructure(cnr).
		WithContainerNodes(nodes).
		WithContainerID(cid)

	if err := ap.taskManager.PushTask(task); err != nil {
		ap.log.Error("could not push audit task",
			zap.String("error", err.Error()),
		)
	}
}

func (ap *Processor) findStorageGroups(cid *container.ID, shuffled netmap.Nodes) []*object.ID {
	var sg []*object.ID

	ln := len(shuffled)

	for i := range shuffled { // consider iterating over some part of container
		log := ap.log.With(
			zap.Stringer("cid", cid),
			zap.String("address", shuffled[0].Address()),
			zap.Int("try", i),
			zap.Int("total_tries", ln),
		)

		ctx, cancel := context.WithTimeout(context.Background(), ap.searchTimeout)
		result, err := ap.sgSearcher.SearchSG(ctx, shuffled[i], cid)
		cancel()

		if err != nil {
			log.Warn("error in storage group search", zap.String("error", err.Error()))
			continue
		}

		sg = append(sg, result...)

		break // we found storage groups, so break loop
	}

	return sg
}
