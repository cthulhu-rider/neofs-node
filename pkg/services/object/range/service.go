package rangesvc

import (
	"context"
	"crypto/ecdsa"
	"sync"

	"github.com/nspcc-dev/neofs-api-go/pkg/object"
	"github.com/nspcc-dev/neofs-node/pkg/core/container"
	"github.com/nspcc-dev/neofs-node/pkg/core/netmap"
	"github.com/nspcc-dev/neofs-node/pkg/local_object_storage/localstore"
	"github.com/nspcc-dev/neofs-node/pkg/network"
	headsvc "github.com/nspcc-dev/neofs-node/pkg/services/object/head"
	"github.com/nspcc-dev/neofs-node/pkg/util"
	"github.com/pkg/errors"
)

type Service struct {
	*cfg
}

type Option func(*cfg)

type cfg struct {
	key *ecdsa.PrivateKey

	localStore *localstore.Storage

	cnrSrc container.Source

	netMapSrc netmap.Source

	workerPool util.WorkerPool

	localAddrSrc network.LocalAddressSource

	headSvc *headsvc.Service
}

func defaultCfg() *cfg {
	return &cfg{
		workerPool: new(util.SyncWorkerPool),
	}
}

func NewService(opts ...Option) *Service {
	c := defaultCfg()

	for i := range opts {
		opts[i](c)
	}

	return &Service{
		cfg: c,
	}
}

func (s *Service) GetRange(ctx context.Context, prm *Prm) (*Result, error) {
	headResult, err := s.headSvc.Head(ctx, new(headsvc.Prm).
		WithAddress(prm.addr).
		OnlyLocal(prm.local),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "(%T) could not receive Head result", s)
	}

	off, ln := prm.rng.GetOffset(), prm.rng.GetLength()

	origin := headResult.Header()

	originSize := origin.GetPayloadSize()
	if originSize < off+ln {
		return nil, errors.Errorf("(%T) requested payload range is out-of-bounds", s)
	}

	right := headResult.RightChild()
	if right == nil {
		right = origin
	}

	rngTraverser := newRangeTraverser(originSize, right, prm.rng)
	if err := s.fillTraverser(ctx, prm, rngTraverser); err != nil {
		return nil, errors.Wrapf(err, "(%T) could not fill range traverser", s)
	}

	return &Result{
		head: origin,
		stream: &streamer{
			cfg:            s.cfg,
			once:           new(sync.Once),
			ctx:            ctx,
			prm:            prm,
			rangeTraverser: rngTraverser,
		},
	}, nil
}

func (s *Service) fillTraverser(ctx context.Context, prm *Prm, traverser *rangeTraverser) error {
	addr := object.NewAddress()
	addr.SetContainerID(prm.addr.GetContainerID())

	for {
		next := traverser.next()
		if next.rng != nil {
			return nil
		}

		addr.SetObjectID(next.id)

		head, err := s.headSvc.Head(ctx, new(headsvc.Prm).
			WithAddress(addr).
			OnlyLocal(prm.local),
		)
		if err != nil {
			return errors.Wrapf(err, "(%T) could not receive object header", s)
		}

		traverser.pushHeader(head.Header())
	}
}

func WithKey(v *ecdsa.PrivateKey) Option {
	return func(c *cfg) {
		c.key = v
	}
}

func WithLocalStorage(v *localstore.Storage) Option {
	return func(c *cfg) {
		c.localStore = v
	}
}

func WithContainerSource(v container.Source) Option {
	return func(c *cfg) {
		c.cnrSrc = v
	}
}

func WithNetworkMapSource(v netmap.Source) Option {
	return func(c *cfg) {
		c.netMapSrc = v
	}
}

func WithWorkerPool(v util.WorkerPool) Option {
	return func(c *cfg) {
		c.workerPool = v
	}
}

func WithLocalAddressSource(v network.LocalAddressSource) Option {
	return func(c *cfg) {
		c.localAddrSrc = v
	}
}

func WithHeadService(v *headsvc.Service) Option {
	return func(c *cfg) {
		c.headSvc = v
	}
}