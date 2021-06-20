package main

import (
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-api-go/pkg/client"
	"github.com/nspcc-dev/neofs-api-go/rpc/message"
	"github.com/nspcc-dev/neofs-api-go/v2/accounting"
	accounting2 "github.com/nspcc-dev/neofs-api-go/v2/accounting/grpc"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc"
	"github.com/nspcc-dev/neofs-api-go/v2/signature"
)

type accountingSvc struct {
	c client.Client

	key *keys.PrivateKey
}

func (x accountingSvc) Balance(w *ReqWrapper) (*RespWrapper, error) {
	req := new(accounting.BalanceRequest)
	req.SetBody(&w.r)

	err := signature.SignServiceMessage(&x.key.PrivateKey, req)
	if err != nil {
		return nil, err
	}

	r, err := rpc.Balance(x.c.Raw(), req)
	if err != nil {
		return nil, err
	}

	return &RespWrapper{*r}, nil
}

type ReqWrapper struct {
	r accounting.BalanceRequestBody
}

func (r ReqWrapper) MarshalJSON() ([]byte, error) {
	return message.MarshalJSON(&r.r)
}

func (r *ReqWrapper) UnmarshalJSON(data []byte) error {
	return message.UnmarshalJSON(&r.r, data, new(accounting2.BalanceRequest_Body))
}

type RespWrapper struct {
	r accounting.BalanceResponse
}

func (r *RespWrapper) MarshalJSON() ([]byte, error) {
	return message.MarshalJSON(&r.r)
}

func (r *RespWrapper) UnmarshalJSON(data []byte) error {
	return message.UnmarshalJSON(&r.r, data, new(accounting2.BalanceResponse))
}
