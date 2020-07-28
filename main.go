package main

import (
	"fmt"
	"log"
	"os"

	"flag"

	"github.com/cashtang/url-shortener/shortener"

	"github.com/gorilla/mux"
)

func serve(a shortener.App) {
	a.Run(fmt.Sprintf(":%v", a.Config.Port))
}

func main() {
	var configFile string

	flag.StringVar(&configFile, "config", "./config.yaml", "config file path, default config.yaml")

	flag.Parse()

	a := shortener.App{}
	r := mux.NewRouter()
	if err := a.Init(configFile, r); err != nil {
		log.Println("init app error, ", err)
		os.Exit(1)
	}
	serve(a)
}
