package main

import (
	copshq "github.com/conplementag/cops-hq/v2/pkg/hq"
	"github.com/conplementag/cops-vigilante/internal/vigilante/config"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	defer errorhandler()

	hq := copshq.NewCustom("cops-vigilante", "1.0.0", &copshq.HqOptions{
		Quiet: false,

		// the service normally runs in a Kubernetes env, so logging to a file is not necessary.
		// Also, in a distro-less image we don't really want to write to the file system anyway.
		DisableFileLogging: true,
		LogFileName:        "",
	})

	// calling this method before anything else is instantiated (like viper CLI parameter connect) so that the order of
	// config overwrites is kept
	config.LoadConfigFiles()

	createCommands(hq)
	hq.Run()
}

func errorhandler() {
	if r := recover(); r != nil {
		logrus.Errorf("Unhandled exception terminating the application: %+v\n", r)
		os.Exit(1)
	}
}
