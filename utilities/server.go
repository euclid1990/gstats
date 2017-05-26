package utilities

import (
	"fmt"
	"github.com/euclid1990/gstats/configs"
	"github.com/gorilla/mux"
	"net/http"
)

var (
	googleOauth *GoogleOauth
	githubOauth *GithubOauth
	googleCodes = make(map[string]bool)
	githubCodes = make(map[string]bool)
)

func Server(google *GoogleOauth, github *GithubOauth) {
	var addr = fmt.Sprintf(":%v", configs.SERVER_PORT)
	googleOauth = google
	githubOauth = github
	r := mux.NewRouter()
	r.HandleFunc(configs.SERVER_GOOGLE_CALLBACK, googleRoute).Methods("GET")
	r.HandleFunc(configs.SERVER_GITHUB_CALLBACK, githubRoute).Methods("GET")
	http.ListenAndServe(addr, r)
}

func googleRoute(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	code, ok := params["code"]
	if ok {
		code := code[0]
		if _, ok := googleCodes[code]; !ok {
			googleOauth.codeChan <- code
			googleCodes[code] = true
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(code))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("[Google Oauth] Unable to read Google authorization code."))
}

func githubRoute(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	code, ok := params["code"]
	if ok {
		code := code[0]
		if _, ok := githubCodes[code]; !ok {
			githubOauth.codeChan <- code
			githubCodes[code] = true
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(code))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("[Github Oauth] Unable to read Google authorization code."))
}
