package util

import (
	"fmt"
	"github.com/spf13/viper"
)

type Conf interface {
	LeafConf
	Sub(key string) *viper.Viper
}

type LeafConf interface {
	GetString(key string) string
	GetInt(key string) int
}

const mongoURLFormatNoAuth = "mongodb://%s:%d/%s?appName=%s"
const mongoURLFormat = "mongodb://%s:%d/%s?appName=%s"

func CreateMongoUrl(dbConf LeafConf, appName string) string {
	user := dbConf.GetString("user")
	password := dbConf.GetString("password")
	if user != "" && password != "" {
		return fmt.Sprintf(mongoURLFormat, user, password,
			dbConf.GetString("host"), dbConf.GetInt("port"),
			dbConf.GetString("database"), appName)
	}
	return fmt.Sprintf(mongoURLFormatNoAuth, dbConf.GetString("host"),
		dbConf.GetInt("port"), dbConf.GetString("database"), appName)
}
