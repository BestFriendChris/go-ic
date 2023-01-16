package cmd

import (
	"flag"
	"os"
)

type Cmd struct {
	fc flagChecker
}

func New() *Cmd {
	return &Cmd{
		fc: &globalFlagChecker{},
	}
}

func NewNullable() (*Cmd, *OverridableFlagChecker) {
	fc := &OverridableFlagChecker{}
	return &Cmd{
		fc: fc,
	}, fc
}

func (c *Cmd) IsUpdateEnabled() bool {
	return c.fc.IsUpdateEnabled()
}

type flagChecker interface {
	IsUpdateEnabled() bool
}

type globalFlagChecker struct{}

func (g *globalFlagChecker) IsUpdateEnabled() bool {
	_, envUpdateEnabled := os.LookupEnv("IC_UPDATE")
	return *updateEnabled || envUpdateEnabled
}

type OverridableFlagChecker struct {
	FlagEnabled, EnvEnabled bool
}

func (o *OverridableFlagChecker) IsUpdateEnabled() bool {
	return o.FlagEnabled || o.EnvEnabled
}

var (
	updateEnabled *bool
)

func init() {
	updateEnabled = flag.Bool("test.icupdate", false, "allow IC to update test files")
}
