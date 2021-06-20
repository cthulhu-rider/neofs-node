package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/nspcc-dev/neofs-api-go/pkg/client"
	"github.com/nspcc-dev/neofs-node/cmd/neofs-node/config"
	"github.com/nspcc-dev/neofs-node/pkg/util/grace"
)

type handler struct {
	srv *rpc.Server
}

func (x handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	x.srv.ServeHTTP(rw, req)
}

func serveGate(c *config.Config) {
	httpEndpoint := listenHTTPEndpoint(c)
	jsonEndpoint := listenJSONEnpoint(c)

	cl, err := client.New(
		client.WithAddress(nodeEndpoint(c)),
	)
	if err != nil {
		panic(fmt.Errorf("could not create NeoFS client: %w", err))
	}

	srv := rpc.NewServer()

	accSvc := &accountingSvc{
		c:   cl,
		key: key(c),
	}

	const accSvcName = "neo.fs.v2"

	err = srv.RegisterName("accounting", accSvc)
	if err != nil {
		panic(fmt.Errorf("register %s service failure: %w", accSvcName, err))
	}

	httpSrv := &http.Server{
		Addr:    httpEndpoint,
		Handler: handler{srv},
	}

	httpLis, err := net.Listen("tcp", httpEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	jsonLis, err := net.Listen("tcp", jsonEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		httpSrv.Close()
		srv.Stop()
		httpLis.Close()
		jsonLis.Close()
	}()

	ctx := grace.NewGracefulContext(nil)

	go func() {
		log.Println("serving HTTP on", httpEndpoint)
		err := httpSrv.Serve(httpLis)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		log.Println("serving JSON RPC on", jsonEndpoint)
		err := srv.ServeListener(jsonLis)
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
}
