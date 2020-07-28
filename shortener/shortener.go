package shortener

import (
	"database/sql"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

const defaultConfig = `
port: 8080
storage: redis
storage_connect_url: 127.0.0.1:5432
public_url: http://localhost:8080
`

// AppConfig application configuration
type AppConfig struct {
	// Port serve http port
	Port int `yaml:"port"`
	// Storage URL storage
	Storage string `yaml:"storage"`
	// StorageConnectURL storage connection url
	StorageConnectURL string `yaml:"storage_connect_url"`
	// PublicURL public url
	PublicURL string `yaml:"public_url"`
}

//App with a router and db as dependencies
type App struct {
	Router *mux.Router
	Config AppConfig
	DB     *sql.DB
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
	return nil
}

//Init routes
func (a *App) Init(configFile string, router *mux.Router) error {
	if err := a.loadConfig(configFile); err != nil {
		return err
	}
	a.Router = router
	a.initRouter()
	return nil
}

//Run the app
func (a *App) Run(port string) {
	log.Fatal(http.ListenAndServe(port, a.Router))
}
