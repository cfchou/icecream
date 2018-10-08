package util

import (
	"fmt"
	"github.com/spf13/viper"
)

// Conf represent a hierarchy of config
type Conf interface {
	LeafConf
	Sub(key string) *viper.Viper
}

// LeafConf reprsents the leaf in a hierarchy of config
type LeafConf interface {
	GetString(key string) string
	GetInt(key string) int
}

const mongoURLFormatNoAuth = "mongodb://%s:%d/%s?appName=%s"
const mongoURLFormat = "mongodb://%s:%s@%s:%d/%s?appName=%s"

// CreateMongoURL reads information verbatim from the config and composes the
// url for connecting to MongoDB. It doesn't validate the config.
func CreateMongoURL(dbConf LeafConf, appName string) string {
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
