package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hauke96/sigolo"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	requestTokenUrl   = "/oauth/request_token"
	authorizeTokenUrl = "/oauth/authorize"
	accessTokenUrl    = "/oauth/access_token"

	registerUserUrl = "/register/{user}"

	redirectUrls = make(map[string]string)
)

func main() {
	router := mux.NewRouter()

	// Oauth Endpoints
	router.HandleFunc(requestTokenUrl, handleRequestToken).Methods(http.MethodPost)
	router.HandleFunc(authorizeTokenUrl, handleAuthorizeToken).Methods(http.MethodGet)
	router.HandleFunc(accessTokenUrl, handleAccessToken).Methods(http.MethodPost)

	// Helper endpoint
	router.HandleFunc(registerUserUrl, handleRegisterUser).Methods(http.MethodGet)

	sigolo.Info("Started router")

	err := http.ListenAndServe(":9000", router)
	sigolo.FatalCheck(err)

	sigolo.Info("Start serving ...")
}

func handleRequestToken(w http.ResponseWriter, r *http.Request) {
	sigolo.Info("Called URL: %#v", r.URL)

	// Read body
	body, err := ioutil.ReadAll(r.Body)
	sigolo.FatalCheck(err)

	// Parse query parameter from body
	queryValues, err := url.ParseQuery(string(body))
	sigolo.FatalCheck(err)

	sigolo.Info("With query values: %#v", queryValues)

	oauthToken := time.Now().Unix()
	result := fmt.Sprintf("&oauth_token=%d&oauth_token_secret=%d&oauth_callback_confirmed=true", oauthToken, time.Now().Unix())

	redirectUrls[fmt.Sprintf("%d", oauthToken)] = queryValues.Get("oauth_callback")

	//redirectTo := fmt.Sprintf("%s%s", queryValues.Get("oauth_callback"), result)
	//sigolo.Info("Redirect to: %s", redirectTo)

	fmt.Fprint(w, result)
	//http.Redirect(w, r, redirectTo, http.StatusMovedPermanently)
}

func handleAuthorizeToken(w http.ResponseWriter, r *http.Request) {
	sigolo.Info("Called URL: %#v", r.URL)

	id := r.URL.Query().Get("oauth_token")
	sigolo.Info("Fill template with ID: %s", id)

	tmpl := template.Must(template.ParseFiles("login.html"))

	// Submit URL: /oauth_callback?redirect=http://localhost:4200/oauth-landing&config=977c642c76c61b4dff79ed9d754087239121f9c2299741952356704bc0358ad4&oauth_token=U9JfSmLqzQIrtBet7bGvihkXKrKh22mJe8BeTSAp&oauth_verifier=du01cYv4qvUyLim7kcgw
	submitUrl := redirectUrls[id] + "&oauth_token=U9JfSmLqzQIrtBet7bGvihkXKrKh22mJe8BeTSAp&oauth_verifier=du01cYv4qvUyLim7kcgw"
	sigolo.Info("Submit URL: %s", submitUrl)

	err := tmpl.Execute(w, struct {
		Id          string
		RedirectUrl string
	}{
		r.URL.Query().Get("oauth_token"),
		submitUrl,
	})
	sigolo.FatalCheck(err)
}

func handleAccessToken(w http.ResponseWriter, r *http.Request) {
	sigolo.Info("Called URL: %#v", r.URL)

	// Read body
	body, err := ioutil.ReadAll(r.Body)
	sigolo.FatalCheck(err)

	sigolo.Info("Body: %s", string(body))

	fmt.Fprint(w, "oauth_token=foo&oauth_verifier=ver&oauth_token_secret=bar")
}

func handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	user := mux.Vars(r)["user"]
	sigolo.Info("Register user: %s", user)
	// TODO register user
}
