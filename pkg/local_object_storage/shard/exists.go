package shard

import (
	"github.com/nspcc-dev/neofs-api-go/pkg/object"
)

// ExistsPrm groups the parameters of Exists operation.
type ExistsPrm struct {
	addr *object.Address
}

// ExistsRes groups resulting values of Exists operation.
type ExistsRes struct {
	ex bool
}

// WithAddress is an Exists option to set object checked for existence.
func (p *ExistsPrm) WithAddress(addr *object.Address) *ExistsPrm {
	if p != nil {
		p.addr = addr
	}

	return p
}

// Exists returns the fact that the object is in the shard.
func (p *ExistsRes) Exists() bool {
	return p.ex
}

// Exists checks if object is presented in shard.
//
// Returns any error encountered that does not allow to
// unambiguously determine the presence of an object.
func (s *Shard) Exists(prm *ExistsPrm) (*ExistsRes, error) {
	// FIXME: implement me
	return &ExistsRes{
		ex: false,
	}, nil
}
