package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hauke96/sigolo"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	requestTokenUrl   = "/oauth/request_token"
	authorizeTokenUrl = "/oauth/authorize"
	accessTokenUrl    = "/oauth/access_token"

	userUrl       = "/api/0.6/user/details"
	changesetsUrl = "/api/0.6/changesets"
	usersUrl      = "/api/0.6/users"

	registerUserUrl = "/register/{oauth_token}/{id}/{user}"

	redirectUrls       = make(map[string]string)
	registeredUsers    = make(map[string]string)

	// This map is used to replace the oauth-token by the user-ID in the "handleAccessToken" function. We need this in
	// order to get the user data for the same user ID in "handleUserData". Otherwise we would get different user data
	// because the oauth token changes on every login.
	oauthTokenToUserId = make(map[string]string)
)

func main() {
	sigolo.Info("Register dummy users")
	registeredUsers["1"] = "john"
	registeredUsers["2"] = "maria"

	router := mux.NewRouter()
	router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,PUT")
		w.Header().Set("Access-Control-Allow-Request-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Request-Methods", "GET,POST,DELETE,PUT")
	})

	// Oauth Endpoints
	router.HandleFunc(requestTokenUrl, handleRequestToken).Methods(http.MethodPost)
	router.HandleFunc(authorizeTokenUrl, handleAuthorizeToken).Methods(http.MethodGet)
	router.HandleFunc(accessTokenUrl, handleAccessToken).Methods(http.MethodPost)

	// OSM API
	router.HandleFunc(userUrl, handleUserData).Methods(http.MethodGet)
	router.HandleFunc(changesetsUrl, handleGetChangeset).Methods(http.MethodGet)
	router.HandleFunc(usersUrl, handleGetUsers).Methods(http.MethodGet)

	// Helper endpoint
	router.HandleFunc(registerUserUrl, handleRegisterUser).Methods(http.MethodGet)

	sigolo.Info("Started router")

	err := http.ListenAndServe(":9000", router)
	sigolo.FatalCheck(err)

	sigolo.Info("Start serving ...")
}

func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	sigolo.Info("Called URL: %#v", r.URL.Path)

	userIDs := strings.Split(r.URL.Query().Get("users"), ",")
	sigolo.Info("Requested users: %#v", userIDs)
	sigolo.Info("Registered users: %#v", registeredUsers)

	users := `<osm version="0.6" generator="OpenStreetMap server" copyright="OpenStreetMap and contributors" attribution="http://www.openstreetmap.org/copyright" license="http://opendatacommons.org/licenses/odbl/1-0/">`

	for _, u := range userIDs {
		user, ok := registeredUsers[u]
		if !ok {
			sigolo.Info("User not found: %s", u)
			user = "<unknown>"
		}

		users += `
<user id="` + u + `" display_name="` + user + `" account_created="2020-05-11T13:43:17Z">
<description/>
<contributor-terms agreed="false"/>
<roles> </roles>
<changesets count="0"/>
<traces count="0"/>
<blocks>
<received count="0" active="0"/>
</blocks>
</user>`
	}

	w.Write([]byte(users + "\n</osm>"))
}

func handleGetChangeset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	sigolo.Info("Called URL: %#v", r.URL.Path)

	userName := r.URL.Query().Get("display_name")

	var uid string
	for i, u := range registeredUsers {
		if u == userName {
			uid = i
		}
	}

	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<osm version="0.6" generator="OpenStreetMap server" copyright="OpenStreetMap and contributors" attribution="http://www.openstreetmap.org/copyright" license="http://opendatacommons.org/licenses/odbl/1-0/">
<changeset id="1" created_at="2020-05-12T12:19:39Z" open="false" comments_count="0" changes_count="5" closed_at="2020-05-12T12:29:39Z" min_lat="53" min_lon="9" max_lat="54" max_lon="10" uid="`+uid+`" user="`+userName+`">
  <tag k="changesets_count" v="1"/>
  <tag k="imagery_used" v="foo"/>
  <tag k="locale" v="en-US"/>
  <tag k="host" v="foo.com"/>
  <tag k="created_by" v="fooEdit 14.3"/>
  <tag k="comment" v="Add foo"/>
</changeset>
</osm>
`)
}

func handleRequestToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	sigolo.Info("Called URL: %#v", r.URL.Path)

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

	fmt.Fprint(w, result)
}

func handleAuthorizeToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	sigolo.Info("Called URL: %#v", r.URL.Path)

	id := r.URL.Query().Get("oauth_token")
	sigolo.Info("Fill template with ID: %s", id)

	tmpl := template.Must(template.ParseFiles("login.html"))

	// Submit URL: /oauth_callback?redirect=http://localhost:4200/oauth-landing&config=977c642c76c61b4dff79ed9d754087239121f9c2299741952356704bc0358ad4&oauth_token=U9JfSmLqzQIrtBet7bGvihkXKrKh22mJe8BeTSAp&oauth_verifier=du01cYv4qvUyLim7kcgw
	submitUrl := redirectUrls[id] + "&foo&oauth_verifier=ver"
	sigolo.Info("Submit URL: %s", submitUrl)

	err := tmpl.Execute(w, struct {
		OauthToken  string
		RedirectUrl string
		Users map[string]string
	}{
		r.URL.Query().Get("oauth_token"),
		submitUrl,
		registeredUsers,
	})
	sigolo.FatalCheck(err)
}

func handleAccessToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	sigolo.Info("Called URL: %#v", r.URL.Path)

	oauthToken := oauthTokenToUserId[getToken(r)]

	// Read body
	body, err := ioutil.ReadAll(r.Body)
	sigolo.FatalCheck(err)

	sigolo.Info("Body: %s", string(body))

	fmt.Fprint(w, "oauth_token="+oauthToken+"&oauth_verifier=ver&oauth_token_secret=bar")
}

func handleUserData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	sigolo.Info("Called URL: %#v", r.URL.Path)

	oauthToken := getToken(r)

	fmt.Fprint(w, `<osm version="0.6" generator="OpenStreetMap server" copyright="OpenStreetMap and contributors" attribution="http://www.openstreetmap.org/copyright" license="http://opendatacommons.org/licenses/odbl/1-0/">
<user id="`+oauthToken+`" display_name="`+registeredUsers[oauthToken]+`" account_created="2020-03-26T22:24:52Z">
  <description></description>
  <contributor-terms agreed="true" pd="false"/>
  <roles>
  </roles>
  <changesets count="1"/>
  <traces count="0"/>
  <blocks>
    <received count="0" active="0"/>
  </blocks>
  <languages>
    <lang>en-US</lang>
    <lang>en</lang>
  </languages>
  <messages>
    <received count="0" unread="0"/>
    <sent count="0"/>
  </messages>
</user>
</osm>`)
}

func handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	sigolo.Info("Called URL: %#v", r.URL.Path)

	id := mux.Vars(r)["id"]
	sigolo.Info("id: %s", id)

	user := mux.Vars(r)["user"]
	sigolo.Info("Register user: %s", user)

	oauthToken := mux.Vars(r)["oauth_token"]
	sigolo.Info("OAuth Token: %s", oauthToken)

	registeredUsers[id] = user
	oauthTokenToUserId[oauthToken] = id

	sigolo.Info("Registering user done: %#v", oauthTokenToUserId)
}

func getToken(r *http.Request) string {
	sigolo.Info("Headers: %#v", r.Header)

	authSegments := strings.Split(r.Header.Get("Authorization"), ", ")
	var oauthToken string
	for _, s := range authSegments {
		if strings.HasPrefix(s, "oauth_token") {
			oauthToken = s[13 : len(s)-1]
		}
	}
	sigolo.Info("Found oauth_token: %s", oauthToken)
	return oauthToken
}
