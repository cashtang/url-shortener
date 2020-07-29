package shortener

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

const defaultConfig = `app:
  port: 8080
  ttl: 4320 # 24 * 180
  storage_connect_url: 127.0.0.1:5432
  public_url: http://localhost:8080
`

// AppConfig application configuration
type AppConfig struct {
	App struct {
		// Port serve http port
		Port int `yaml:"port"`
		// TTL in second
		TTL int `yaml:"ttl"`
		// StorageConnectURL storage connection url
		StorageConnectURL string `yaml:"storage_connect_url"`
		// PublicURL public url
		PublicURL string `yaml:"public_url"`
	}
}

//App with a router and db as dependencies
type App struct {
	Router  *mux.Router
	Config  AppConfig
	Storage URLStorage
}

func (a *App) initRouter() {
	a.Router.HandleFunc("/", a.Home).Methods("GET")
	a.Router.HandleFunc("/{hash}", a.Redirect).Methods("GET")
	a.Router.HandleFunc("/shorten", a.Shorten).Methods("POST")
}

func (a *App) loadConfig(configFile string) error {
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println("load config ", configFile, " , error ", err)
		return err
	}
	err = yaml.Unmarshal(content, &a.Config)
	if err != nil {
		log.Println("load config ", configFile, " , error ", err)
		return err
	}
	if a.Config.App.TTL <= 0 {
		a.Config.App.TTL = 6000
	}
	return nil
}

//Init routes
func (a *App) Init(configFile string, router *mux.Router) error {
	if err := a.loadConfig(configFile); err != nil {
		return err
	}
	a.Router = router
	a.initRouter()

	if r, err := InitStorage(a.Config.App.StorageConnectURL); err == nil {
		a.Storage = r
	}
	return nil
}

// GenerateConfig generate config file template
func (a *App) GenerateConfig(configFile string) {
	if _, err := os.Stat(configFile); err == nil {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Config file <", configFile, "> already exist, overwrite (Y/n)?")
		answer, _ := reader.ReadString('\n')
		answer = strings.ToLower(strings.TrimRight(answer, "\n"))
		// fmt.Printf("Answer is <%v>\n", answer)
		switch {
		case answer == "y":
			break
		case answer == "n":
			fmt.Println("Generate config file canceled!")
			return
		default:
			break
		}
	}
	ioutil.WriteFile(configFile, []byte(defaultConfig), 0660)
	log.Println("Generate config success,", configFile)
}

//Run the app
func (a *App) Run(port string) {
	log.Fatal(http.ListenAndServe(port, a.Router))
}
