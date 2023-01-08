package main

import (
	copshq "github.com/conplementag/cops-hq/v2/pkg/hq"
	"github.com/conplementag/cops-vigilante/internal/vigilante"
	"github.com/conplementag/cops-vigilante/internal/vigilante/cli"
)

func createCommands(hq copshq.HQ) {
	runCommand := hq.GetCli().AddBaseCommand("run", "start the vigilante application", "start the vigilante application", func() {
		vigilante.Run()
	})

	runCommand.AddParameterBool(cli.TLSFlag, false, false, "t", "Set to load the TLS certificate from the /etc/certs/tls.crt and /etc/certs/tls.key locations.")
	runCommand.AddParameterInt(cli.IntervalInSecondsFlag, 30, false, "i", "Interval in seconds for the scheduler loop which executes the healing logic.")
}
