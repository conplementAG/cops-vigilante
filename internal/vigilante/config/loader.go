package config

import (
	copshq "github.com/conplementag/cops-hq/v2/pkg/hq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

// LoadConfigFiles loads the local disk config files, with the order of overrides (first the /etc config, then the local config)
func LoadConfigFiles() error {
	viper.SetConfigType("yaml")

	err := loadFile(filepath.Join("/", "etc", "vigilante", "config", "conf.yaml"))

	if err != nil {
		return err
	}

	return loadFile(filepath.Join(copshq.ProjectBasePath, "config", "conf.yaml"))
}

func loadFile(filePath string) error {
	viper.SetConfigType("yaml")

	fileReader, err := os.Open(filePath)

	if err == nil {
		err = viper.MergeConfig(fileReader)

		logrus.Info("Config file found and loaded from location " + filePath)

		if err != nil {
			logrus.Error("Error parsing ")
			return err
		}
	} else {
		logrus.Info("No config file found in " + filePath)
		logrus.Debugf("Error recieved when search for the file %s was %v", filePath, err)
	}

	return nil
}
