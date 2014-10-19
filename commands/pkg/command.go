package pkgCmd

import (
	"github.com/salsita/salsaflow/commands/pkg/install"
	"github.com/salsita/salsaflow/commands/pkg/upgrade"

	"gopkg.in/tchap/gocli.v1"
)

var Command = &gocli.Command{
	UsageLine: "pkg",
	Short:     "manage SalsaFlow executables",
	Long: `
  Manage SalsaFlow executables. See the subcommands.
	`,
}

func init() {
	Command.MustRegisterSubcommand(installCmd.Command)
	Command.MustRegisterSubcommand(upgradeCmd.Command)
}