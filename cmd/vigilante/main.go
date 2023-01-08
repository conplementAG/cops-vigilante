package main

import (
	copshq "github.com/conplementag/cops-hq/v2/pkg/hq"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	defer errorhandler()

	hq := copshq.NewCustom("cops-vigilante", "0.0.1", &copshq.HqOptions{
		Quiet: false,

		// the service normally runs in a Kubernetes env, so logging to a file is not necessary.
		// Also, in a distroless image we don't really want to write to the file system anyway.
		DisableFileLogging: true,
		LogFileName:        "",
	})

	createCommands(hq)
	hq.Run()
}

func errorhandler() {
	if r := recover(); r != nil {
		logrus.Errorf("Unhandled exception terminating the application: %+v\n", r)
		os.Exit(1)
	}
}
