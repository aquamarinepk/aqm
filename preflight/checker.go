package preflight

import (
	"context"
	"fmt"

	"github.com/aquamarinepk/aqm/log"
)

type Check interface {
	Name() string
	Run(ctx context.Context) error
}

type Checker struct {
	checks []Check
	log    log.Logger
}

func New(logger log.Logger) *Checker {
	return &Checker{
		checks: make([]Check, 0),
		log:    logger,
	}
}

func (c *Checker) Add(check Check) *Checker {
	c.checks = append(c.checks, check)
	return c
}

func (c *Checker) RunAll(ctx context.Context) error {
	if len(c.checks) == 0 {
		c.log.Debugf("No preflight checks configured")
		return nil
	}

	c.log.Infof("Running %d preflight checks", len(c.checks))

	for _, check := range c.checks {
		c.log.Debugf("Running preflight check: %s", check.Name())

		if err := check.Run(ctx); err != nil {
			c.log.Errorf("Preflight check failed: %s - %v", check.Name(), err)
			return fmt.Errorf("preflight check %q failed: %w", check.Name(), err)
		}

		c.log.Infof("Preflight check passed: %s", check.Name())
	}

	c.log.Infof("All preflight checks passed")
	return nil
}
