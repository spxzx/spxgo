package config

import (
	"flag"
	spxLog "gitbuh.com/spxzx/spxgo/log"
	"github.com/BurntSushi/toml"
	"os"
)

var Conf = &SpxConfig{
	logger: spxLog.Default(),
}

type SpxConfig struct {
	logger *spxLog.Logger
	Log    map[string]any
	Pool   map[string]any
	Mysql  map[string]any
}

func init() {
	loadToml()
}

func loadToml() {
	configFile := flag.String("conf", "conf/app.toml", "app default config file")
	flag.Parse()
	if _, err := os.Stat(*configFile); err != nil {
		Conf.logger.Info("conf/app.toml file not load, because not exist")
		return
	}
	_, err := toml.DecodeFile(*configFile, Conf)
	if err != nil {
		Conf.logger.Error("conf/app.toml decode fail, please check format")
		return
	}
}
