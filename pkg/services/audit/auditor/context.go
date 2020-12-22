package auditor

import (
	"github.com/nspcc-dev/neofs-api-go/pkg/container"
	"github.com/nspcc-dev/neofs-node/pkg/services/audit"
	"github.com/nspcc-dev/neofs-node/pkg/util/logger"
	"go.uber.org/zap"
)

// Context represents container data audit execution context.
type Context struct {
	ContextPrm

	task *audit.Task

	report *audit.Report
}

// ContextPrm groups components required to conduct data audit checks.
type ContextPrm struct {
	log *logger.Logger
}

// NewContext creates, initializes and returns Context.
func NewContext(prm ContextPrm) *Context {
	return &Context{
		ContextPrm: prm,
	}
}

// SetLogger sets logging component.
func (p *ContextPrm) SetLogger(l *logger.Logger) {
	if p != nil {
		p.log = l
	}
}

// WithTask sets container audit parameters.
func (c *Context) WithTask(t *audit.Task) *Context {
	if c != nil {
		c.task = t
	}

	return c
}

func (c *Context) containerID() *container.ID {
	return c.task.ContainerID()
}

func (c *Context) init() {
	c.report = audit.NewReport(c.containerID())

	c.log = c.log.With(
		zap.Stringer("container ID", c.task.ContainerID()),
	)
}

func (c *Context) expired() bool {
	ctx := c.task.AuditContext()

	select {
	case <-ctx.Done():
		c.log.Debug("audit context is done",
			zap.String("error", ctx.Err().Error()),
		)

		return true
	default:
		return false
	}
}

func (c *Context) writeReport() {
	c.log.Debug("writing audit report...")

	if err := c.task.Reporter().WriteReport(c.report); err != nil {
		c.log.Error("could not write audit report")
	}
}
