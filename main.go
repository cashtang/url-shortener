package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"
	"time"

	"flag"

	"github.com/cashtang/url-shortener/shortener"

	"github.com/gorilla/mux"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
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

func setupLog() {
	var logDir = ""
	if runtime.GOOS == "windows" {
		logDir = os.TempDir()
	} else {
		logDir = "/var/log/url-shortener"
	}
	logFile := path.Join(logDir, "url-shortener.log")
	logf, err := rotatelogs.New(
		fmt.Sprintf("%v.%v", logFile, "%Y%m%d"),
		rotatelogs.WithLinkName(logFile),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		log.Printf("failed to create rotatelogs: %s", err)
		return
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logf))
	log.Println("setup log success ...", logFile)
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
	a := shortener.App{
		BaseDir: path.Dir(os.Args[0]),
	}
	if generateConfig {
		a.GenerateConfig(configFile)
		return
	}

	setupCloseHandler()
	setupLog()
	r := mux.NewRouter()
	if err := a.Init(configFile, r); err != nil {
		log.Println("init app error, ", err)
		os.Exit(1)
	}
	serve(a)
}
