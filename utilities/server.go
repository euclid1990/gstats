package utilities

import (
	"fmt"
	"github.com/euclid1990/gstats/configs"
	"github.com/gorilla/mux"
	"net/http"
)

var (
	googleOauth *GoogleOauth
	googleCodes = make(map[string]bool)
)

func Server(google *GoogleOauth) {
	var addr = fmt.Sprintf(":%v", configs.SERVER_PORT)
	googleOauth = google
	r := mux.NewRouter()
	r.HandleFunc(configs.SERVER_GOOGLE_CALLBACK, googleRoute).Methods("GET")
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
