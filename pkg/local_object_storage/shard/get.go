package shard

import (
	"errors"

	objectSDK "github.com/nspcc-dev/neofs-api-go/pkg/object"
	"github.com/nspcc-dev/neofs-node/pkg/core/object"
)

// GetPrm groups the parameters of Get operation.
type GetPrm struct {
	ln int64 // negative value for full range

	off uint64

	addr *objectSDK.Address
}

// GetRes groups resulting values of Get operation.
type GetRes struct {
	obj *object.Object
}

// ErrObjectNotFound is returns on read operations requested on a missing object.
var ErrObjectNotFound = errors.New("object not found")

// WithAddress is a Get option to set the address of the requested object.
//
// Option is required.
func (p *GetPrm) WithAddress(addr *objectSDK.Address) *GetPrm {
	if p != nil {
		p.addr = addr
	}

	return p
}

// WithFullRange is a Get option to receive full object payload.
func (p *GetPrm) WithFullRange() *GetPrm {
	if p != nil {
		p.ln = -1
	}

	return p
}

// WithRange is a Get option to set range of requested payload data.
//
// Calling with negative length is equivalent
// to getting the full payload range.
func (p *GetPrm) WithRange(off uint64, ln int64) *GetPrm {
	if p != nil {
		p.off, p.ln = off, ln
	}

	return p
}

// Object returns the requested object part.
//
// Instance payload contains the requested range of the original object.
func (r *GetRes) Object() *object.Object {
	return r.obj
}

// Get reads part of an object from shard.
//
// Returns any error encountered that
// did not allow to completely read the object part.
//
// Returns ErrObjectNotFound if requested object is missing in shard.
func (s *Shard) Get(prm *GetPrm) (*GetRes, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	// FIXME: implement me

	return nil, nil
}