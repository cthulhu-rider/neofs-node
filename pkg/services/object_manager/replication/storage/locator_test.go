package storage

import (
	"testing"

	"github.com/nspcc-dev/neofs-node/pkg/services/object_manager/transport"
	"github.com/nspcc-dev/neofs-node/pkg/util/logger"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type testExecutor struct {
	transport.SelectiveContainerExecutor
}

func TestNewObjectLocator(t *testing.T) {
	validParams := LocatorParams{
		SelectiveContainerExecutor: new(testExecutor),
		Logger:                     zap.L(),
	}

	t.Run("valid params", func(t *testing.T) {
		s, err := NewObjectLocator(validParams)
		require.NoError(t, err)
		require.NotNil(t, s)
	})
	t.Run("empty logger", func(t *testing.T) {
		p := validParams
		p.Logger = nil
		_, err := NewObjectLocator(p)
		require.EqualError(t, err, errors.Wrap(logger.ErrNilLogger, locatorInstanceFailMsg).Error())
	})
	t.Run("empty container handler", func(t *testing.T) {
		p := validParams
		p.SelectiveContainerExecutor = nil
		_, err := NewObjectLocator(p)
		require.EqualError(t, err, errors.Wrap(errEmptyObjectsContainerHandler, locatorInstanceFailMsg).Error())
	})
}
