package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flag"

	"github.com/cashtang/url-shortener/shortener"

	"github.com/gorilla/mux"
)

// VERSION service version
var VERSION = "develop"

// GOVERSION go version
var GOVERSION = "unknown"

func serve(a shortener.App) {
	a.Run(fmt.Sprintf(":%v", a.Config.App.Port))
}

func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("\r - Ctrl-c terminate process!!")
		os.Exit(1)
	}()
}

func main() {
	var configFile string
	var generateConfig, version bool

	flag.StringVar(&configFile, "config", "./config.yaml", "config file path, default config.yaml")
	flag.BoolVar(&generateConfig, "generate", false, "generate config file template")
	flag.BoolVar(&version, "version", false, "show version")

	flag.Parse()

	log.Println("Start:", time.Now())
	if version {
		log.Println("Version:", VERSION)
		log.Println("Build:", GOVERSION)
		return
	}
	a := shortener.App{}
	if generateConfig {
		a.GenerateConfig(configFile)
		return
	}

	setupCloseHandler()
	r := mux.NewRouter()
	if err := a.Init(configFile, r); err != nil {
		log.Println("init app error, ", err)
		os.Exit(1)
	}
	serve(a)
}
