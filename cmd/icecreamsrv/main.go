package main

import (
	"fmt"
	"github.com/inconshreveable/log15"
	"github.com/spf13/viper"
)

const AppName = "Ice cream service"

var log = log15.New("module", "cmd.icecreamsrv")

func init() {
	viper.SetConfigName("icecream")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
}

func main() {
	defer log.Info(fmt.Sprintf("%s stops", AppName))
	log.Info(fmt.Sprintf("%s starts", AppName))
}
