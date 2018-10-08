package main

import (
	"bitbucket.org/cfchou/icecream/cmd/apiserver/handler"
	"bitbucket.org/cfchou/icecream/cmd/apiserver/util"
	"bitbucket.org/cfchou/icecream/pkg/backend/mongodb"
	"bytes"
	"fmt"
	"github.com/inconshreveable/log15"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
)

const appName string = "cmd.apiserver"
const apiVersion string = "v1"

var (
	log           = log15.New("module", appName)
	defaultConfig = []byte(`
server:
  host: 127.0.0.1
  port: 8080
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

	dbConf := viper.Sub("db")
	url := util.CreateMongoUrl(dbConf, appName)
	backend, err := mongodb.CreateMongoBackend(url)

	if err != nil {
		log.Error("CreateMongoBackend failed", "err", err.Error())
		return
	}
	defer backend.Close()

	ph := handler.CreateProductHandler(backend, backend)
	mux := httprouter.New()
	mux.GET("/products/:productID", httprouter.Handle(ph.HandleGet))
	mux.GET("/products", httprouter.Handle(ph.HandleGet))
	mux.POST("/products", httprouter.Handle(ph.HandlePost))
	mux.PUT("/products/:productID", httprouter.Handle(ph.HandlePut))
	mux.DELETE("/products/:productID", httprouter.Handle(ph.HandleDelete))

	serverConf := viper.Sub("server")
	server := http.Server{
		Addr:    fmt.Sprintf("%s:%d", serverConf.GetString("host"),
			serverConf.GetInt("port")),
		Handler: mux,
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

func hello(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", p.ByName("name"))
}

func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return fn
}
