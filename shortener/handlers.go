package shortener

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/gorilla/mux"
	"github.com/teris-io/shortid"
)

const MAX_SIGN_FILE = 1024 * 10

//Body is the response body
type Body struct {
	URL string `json:"url"`
}

// AppBody -
type AppBody struct {
	Secret string `json:"secret"`
}

//Home page
func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Nothing to see here"))
}

func (a *App) verifyAdminToken(w http.ResponseWriter, r *http.Request) bool {
	apiKey := r.Header.Get("ApiKey")
	if apiKey == a.Config.App.AdminToken {
		return true
	}
	respondWithError(w, http.StatusForbidden, "token error")
	return false
}

func (a *App) verifyAppSecret(w http.ResponseWriter, r *http.Request) (string, error) {
	var err error
	var appid string
	key := r.Header.Get("ApiKey")
	appid, err = a.Storage.VerifySecret(key)
	if err != nil {
		if err == ErrAppIDNotFound {
			respondWithError(w, http.StatusForbidden, "appid not exists!")
		} else {
			message := fmt.Sprintf("system error, %v", err)
			respondWithError(w, http.StatusInternalServerError, message)
		}
		return "", err
	}
	return appid, nil
}

//Shorten url POST method
func (a *App) Shorten(w http.ResponseWriter, r *http.Request) {
	var id, appid string
	var body Body
	var err error

	appid, err = a.verifyAppSecret(w, r)
	if err != nil {
		return
	}

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
	if err = a.Storage.NewURL(url, id, appid, a.Config.App.TTL); err != nil {
		message := fmt.Sprintf("Save URL error,%v ", err)
		respondWithError(w, http.StatusInternalServerError, message)
		return
	}
	body.URL = a.Config.App.PublicURL + "/" + id
	sendResponse(w, http.StatusOK, body)
}

// Register -
func (a *App) Register(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appid := vars["appid"]
	secret, err := a.Storage.RegisterAppID(appid)
	if err != nil {
		message := fmt.Sprintf("Register appid error <%v.", err)
		respondWithError(w, http.StatusInternalServerError, message)
		return
	}
	body := &AppBody{}
	body.Secret = secret
	sendResponse(w, http.StatusOK, body)
}

// UnRegister -
func (a *App) UnRegister(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if !a.verifyAdminToken(w, r) {
		return
	}
	appid := vars["appid"]
	err := a.Storage.UnregisterAppID(appid)
	if err != nil {
		message := fmt.Sprintf("Register appid error <%v.", err)
		respondWithError(w, http.StatusInternalServerError, message)
		return
	}
	w.Write([]byte("unregister ok!"))
}

//Redirect route
func (a *App) Redirect(w http.ResponseWriter, r *http.Request) {
	var meta *URLMeta
	var err error

	vars := mux.Vars(r)
	hash := vars["hash"]

	meta, err = a.Storage.FindByID(hash)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Short ID not found")
		return
	}

	u, err := url.ParseRequestURI(r.RequestURI)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Short URL error")
		return
	}
	for k := range u.Query() {
		for _, p := range a.Config.App.ParamsDeny {
			if k == p {
				respondWithError(w, http.StatusBadRequest, "Short Query Param deny")
				return
			}
		}
	}
	params := u.Query().Encode()
	var longURL string
	if len(params) > 0 {
		if strings.Index(meta.LongURL, "?") != -1 {
			longURL = meta.LongURL + "&" + params
		} else {
			longURL = meta.LongURL + "?" + params
		}
	} else {
		longURL = meta.LongURL
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

func (a *App) getSignFullPath(signFile string) string {
	var fullPath string
	if path.IsAbs(a.Config.App.SignFileDir) {
		fullPath = path.Join(a.Config.App.SignFileDir, signFile)
	} else {
		fullPath = path.Join(a.BaseDir, a.Config.App.SignFileDir, signFile)
	}
	return fullPath
}

func (a *App) signedFileGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	signFile := vars["sign"]
	fullPath := a.getSignFullPath(signFile)

	file, err := os.Open(fullPath)
	if err != nil || os.IsNotExist(err) {
		respondWithError(w, http.StatusNotFound, "no file found")
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "read file error")
	}
	w.Write(content)
}

func (a *App) signedFilePost(w http.ResponseWriter, r *http.Request) {
	if !a.verifyAdminToken(w, r) {
		return
	}

	if err := r.ParseMultipartForm(MAX_SIGN_FILE); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	multpart, fileHeader, err := r.FormFile("sign")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer multpart.Close()

	if fileHeader.Size >= MAX_SIGN_FILE {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Println("sign file :", fileHeader.Filename)
	fullPath := a.getSignFullPath(fileHeader.Filename)
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer file.Close()
	if _, err := io.Copy(file, multpart); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Write([]byte("success"))
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
