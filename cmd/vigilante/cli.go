package main

import (
	copshq "github.com/conplementag/cops-hq/v2/pkg/hq"
	"github.com/conplementag/cops-vigilante/internal/vigilante"
	"github.com/conplementag/cops-vigilante/internal/vigilante/cli"
	"github.com/spf13/viper"
)

func createCommands(hq copshq.HQ) {
	runCommand := hq.GetCli().AddBaseCommand("run", "start the vigilante application", "start the vigilante application", func() {
		if viper.GetInt(cli.IntervalFlag) < 15 {
			panic(cli.IntervalFlag + " should not be set to less than 15 seconds. We do not want to flood the k8s API server.")
		}

		vigilante.Run()
	})

	runCommand.AddParameterBool(cli.TLSFlag, false, false, "t", "Set to load the TLS certificate from the /etc/certs/tls.crt and /etc/certs/tls.key locations.")
	runCommand.AddParameterInt(cli.IntervalFlag, 30, false, "i", "Interval in seconds for the scheduler loop which executes all the tasks.")
}
