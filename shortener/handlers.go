package shortener

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/teris-io/shortid"
)

//Body is the response body
type Body struct {
	URL string `json:"url"`
}

//Home page
func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Nothing to see here"))
}

//Shorten url POST method
func (a *App) Shorten(w http.ResponseWriter, r *http.Request) {
	var id string
	var body Body
	var err error

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	url := body.URL

	if !isValidURL(url) {
		respondWithError(w, http.StatusBadRequest, "Invalid url")
		return
	}

	id, err = shortid.Generate()
	if err = a.Storage.NewURL(url, id, a.Config.App.TTL); err != nil {
		message := fmt.Sprintf("Save URL error,%v ", err)
		respondWithError(w, http.StatusInternalServerError, message)
		return
	}
	body.URL = a.Config.App.PublicURL + "/" + id
	sendResponse(w, http.StatusOK, body)
}

//Redirect route
func (a *App) Redirect(w http.ResponseWriter, r *http.Request) {
	var longURL string
	var err error

	vars := mux.Vars(r)
	hash := vars["hash"]

	longURL, err = a.Storage.FindByID(hash)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Short ID not found")
		return
	}

	http.Redirect(w, r, longURL, http.StatusSeeOther)
}

func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}
	return true

}

func respondWithError(w http.ResponseWriter, code int, message string) {
	sendResponse(w, code, map[string]string{"error": message})
}

func sendResponse(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
