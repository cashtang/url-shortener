package shortener

import (
	"encoding/json"
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

	err := a.DB.QueryRow("SELECT id FROM shortened_urls WHERE long_url = ?", url).Scan(&id)
	if err != nil {
		if id, err = shortid.Generate(); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Generate SHORTID error")
			return
		}
		_, err := a.DB.Exec(`INSERT INTO shortened_urls (id, long_url, created) VALUES(?, ?, now())`,
			id, url)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	body.URL = a.Config.PublicURL + "/" + id
	sendResponse(w, http.StatusOK, body)
}

//Redirect route
func (a *App) Redirect(w http.ResponseWriter, r *http.Request) {
	var id string
	var longURL string

	vars := mux.Vars(r)
	hash := vars["hash"]

	err := a.DB.QueryRow("SELECT id, long_url FROM shortened_urls WHERE id = ?", hash).
		Scan(&id, &longURL)
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
