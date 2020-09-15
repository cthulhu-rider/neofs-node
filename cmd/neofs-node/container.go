package main

import (
	containerGRPC "github.com/nspcc-dev/neofs-api-go/v2/container/grpc"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-node/pkg/morph/client"
	"github.com/nspcc-dev/neofs-node/pkg/morph/client/container"
	containerTransportGRPC "github.com/nspcc-dev/neofs-node/pkg/network/transport/container/grpc"
	containerService "github.com/nspcc-dev/neofs-node/pkg/services/container"
	containerMorph "github.com/nspcc-dev/neofs-node/pkg/services/container/morph"
)

func initContainerService(c *cfg) {
	staticClient, err := client.NewStatic(
		c.cfgMorph.client,
		c.cfgContainer.scriptHash,
		c.cfgContainer.fee,
	)
	fatalOnErr(err)

	cnrClient, err := container.New(staticClient)
	fatalOnErr(err)

	metaHdr := new(session.ResponseMetaHeader)
	xHdr := new(session.XHeader)
	xHdr.SetKey("test X-Header key")
	xHdr.SetValue("test X-Header value")
	metaHdr.SetXHeaders([]*session.XHeader{xHdr})

	containerGRPC.RegisterContainerServiceServer(c.cfgGRPC.server,
		containerTransportGRPC.New(
			containerService.NewSignService(
				c.key,
				containerService.NewExecutionService(
					containerMorph.NewExecutor(cnrClient),
					metaHdr,
				),
			),
		),
	)
}