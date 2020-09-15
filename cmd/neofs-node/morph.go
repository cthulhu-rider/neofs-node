package main

import (
	crypto "github.com/nspcc-dev/neofs-crypto"
	"github.com/nspcc-dev/neofs-node/pkg/morph/client"
	"github.com/nspcc-dev/neofs-node/pkg/morph/client/netmap"
	"github.com/nspcc-dev/neofs-node/pkg/morph/client/netmap/wrapper"
	"github.com/pkg/errors"
)

func initMorphComponents(c *cfg) {
	var err error

	c.cfgMorph.client, err = client.New(c.key, c.viper.GetString(cfgMorphRPCAddress))
	fatalOnErr(err)
}

func bootstrapNode(c *cfg) {
	if c.cfgNodeInfo.bootType == StorageNode {
		staticClient, err := client.NewStatic(
			c.cfgMorph.client,
			c.cfgNetmap.scriptHash,
			c.cfgContainer.fee,
		)
		fatalOnErr(err)

		cli, err := netmap.New(staticClient)
		fatalOnErr(errors.Wrap(err, "bootstrap error"))

		cliWrapper, err := wrapper.New(cli)
		fatalOnErr(errors.Wrap(err, "bootstrap error"))

		peerInfo := new(netmap.NodeInfo)
		peerInfo.SetAddress(c.viper.GetString(cfgBootstrapAddress))
		peerInfo.SetPublicKey(crypto.MarshalPublicKey(&c.key.PublicKey))
		// todo: add attributes as opts

		err = cliWrapper.AddPeer(peerInfo)
		fatalOnErr(errors.Wrap(err, "bootstrap error"))
	}
}