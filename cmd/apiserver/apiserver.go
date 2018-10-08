package main

import (
	"bytes"
	"fmt"
	"github.com/cfchou/icecream/cmd/apiserver/handler"
	"github.com/cfchou/icecream/cmd/apiserver/middleware"
	"github.com/cfchou/icecream/cmd/apiserver/util"
	"github.com/cfchou/icecream/pkg/backend/mongodb"
	"github.com/globalsign/mgo"
	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/justinas/alice"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
)

const appName string = "apiserver"

var (
	log           = log15.New("module", appName)
	defaultConfig = []byte(`
server:
  host: 127.0.0.1
  port: 8080
  limitToRead: 10
db:
  database: icecream
  host: 127.0.0.1
  port: 27017
`)
)

func init() {
	var configPath string
	pflag.StringVarP(&configPath, "config-path", "c", "",
		"config(yaml) full path")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(bytes.NewBuffer(defaultConfig)); err != nil {
		fmt.Printf("viper.ReadConfig failed to read the default, %s", err.Error())
		os.Exit(-1)
	}

	// merge: configPath overwrites defaultConfig
	if configPath != "" {
		bs, err := ioutil.ReadFile(configPath)
		if err != nil {
			fmt.Printf("ioutil.ReadFile failed to read %s, %s",
				configPath, err.Error())
			os.Exit(-1)
		}
		if err := viper.MergeConfig(bytes.NewBuffer(bs)); err != nil {
			fmt.Printf("viper.ReadConfig failed to read %s, %s",
				configPath, err.Error())
			os.Exit(-1)
		}
	}
}

func main() {
	defer log.Info(fmt.Sprintf("%s stops", appName))
	log.Info(fmt.Sprintf("%s starts", appName))

	serverConf := viper.Sub("server")
	dbConf := viper.Sub("db")

	url := util.CreateMongoURL(dbConf, appName)
	session, err := mgo.Dial(url)
	if err != nil {
		log.Error("mgo.Dial failed", "err", err.Error())
		return
	}
	defer session.Close()

	productBackend, _ := mongodb.CreateMongoProductBackend(session)
	apiKeyBackend, _ := mongodb.CreateMongoAPIKeyBackend(session)

	ph := handler.CreateProductHandler(productBackend,
		serverConf.GetInt("limitToRead"))

	am := middleware.CreateAPIKeyMiddleWare(apiKeyBackend)
	r := mux.NewRouter()

	// Read
	r.Methods("GET").Path("/products/{productID}").HandlerFunc(ph.HandleGet)
	// Read many, with optional parameters "cursor" and "limit"
	r.Methods("GET").Path("/products/").HandlerFunc(ph.HandleGetMany)

	// Create exclusively without productID in URI. However, the productID must be
	// contained in the payload and must not exsited in the DB.
	r.Methods("POST").Path("/products/").HandlerFunc(ph.HandlePost)

	// Create or replace(fully update).
	r.Methods("PUT").Path("/products/{productID}").HandlerFunc(ph.HandlePut)

	// Partial update
	r.Methods("PATCH").Path("/products/{productID}").HandlerFunc(ph.HandlePatch)

	// Delete
	r.Methods("DELETE").Path("/products/{productID}").HandlerFunc(ph.HandleDelete)

	// Chain middlewares and handler
	stack := alice.New(am.Handle).Then(r)

	server := http.Server{
		Addr: fmt.Sprintf("%s:%d", serverConf.GetString("host"),
			serverConf.GetInt("port")),
		Handler: stack,
	}

	cert := serverConf.GetString("cert")
	key := serverConf.GetString("key")
	if cert != "" && key != "" {
		server.ListenAndServeTLS(cert, key)
	} else {
		log.Warn("Running without SSL")
		server.ListenAndServe()
	}
}
