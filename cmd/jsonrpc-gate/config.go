package main

import (
	"fmt"
	"io/ioutil"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-node/cmd/neofs-node/config"
	config2 "github.com/nspcc-dev/neofs-node/pkg/util/config"
)

func listenHTTPEndpoint(c *config.Config) string {
	v := config.String(c, "http")
	if v == "" {
		panic("missing HTTP server endpoint")
	}

	return v
}

func listenJSONEnpoint(c *config.Config) string {
	v := config.String(c, "jsonrpc")
	if v == "" {
		panic("missing JSON RPC server endpoint")
	}

	return v
}

func nodeEndpoint(c *config.Config) string {
	v := config.String(c, "neofs")
	if v == "" {
		panic("missing NeoFS endpoint")
	}

	return v
}

func key(c *config.Config) *keys.PrivateKey {
	v := config.StringSafe(c, "key")
	if v == "" {
		panic("missing private key")
	}

	var (
		key  *keys.PrivateKey
		err  error
		data []byte
	)
	if data, err = ioutil.ReadFile(v); err == nil {
		key, err = keys.NewPrivateKeyFromBytes(data)
	}

	if err != nil {
		return wallet(c)
	}

	return key
}

func wallet(c *config.Config) *keys.PrivateKey {
	v := c.Sub("wallet")
	acc, err := config2.LoadAccount(
		config.String(v, "path"),
		config.String(v, "address"),
		config.String(v, "password"))
	if err != nil {
		panic(fmt.Errorf("invalid wallet config: %w", err))
	}

	return acc.PrivateKey()
}
