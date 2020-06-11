package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hauke96/sigolo"
	"html/template"
	"net/http"
	"time"
)

const (
	requestTokenUrl   = "/oauth/request_token"
	authorizeTokenUrl = "/oauth/authorize"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", getFrontPage).Methods(http.MethodGet)
	router.HandleFunc(requestTokenUrl, handleRequestToken).Methods(http.MethodPost)
	router.HandleFunc(authorizeTokenUrl, handleAuthorizeToken).Methods(http.MethodGet)

	sigolo.Info("Started router")

	err := http.ListenAndServe(":9000", router)
	sigolo.FatalCheck(err)

	sigolo.Info("Start serving ...")
}

func getFrontPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello world")
}

func handleRequestToken(w http.ResponseWriter, r *http.Request) {
	sigolo.Info("%s called", requestTokenUrl)

	result := fmt.Sprintf("oauth_token=%d&oauth_token_secret=%d&oauth_callback_confirmed=true", time.Now().Unix(), time.Now().Unix())
	sigolo.Info("Return: %s", result)

	fmt.Fprint(w, result)
}

func handleAuthorizeToken(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("oauth_token")
	sigolo.Info("Fill template with ID: %s", id)

	tmpl := template.Must(template.ParseFiles("login.html"))
	err := tmpl.Execute(w, struct{ Id string }{r.URL.Query().Get("oauth_token")})
	sigolo.FatalCheck(err)
}
