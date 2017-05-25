package utilities

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"os"
	"context"
)

const (
	defaultConfigFile = "./private/github_secret.json"
	defaultOAuthFile ="./private/github_oauth.json"

	githubAuthorizeUrl = "https://github.com/login/oauth/authorize"
	githubTokenUrl     = "https://github.com/login/oauth/access_token"
	redirectUrl        = ""
	ACCESS_TOKEN       = "githubAccessToken"
	USERNAME           = "githubUsername"
	STATE              = "githubState"
	SESSION            = "githubSess"
)

type Config struct {
	ClientSecret string `json:"client_secret"`
	ClientID     string `json:"client_id"`
	Secret       string `json:"secret"`
}

type GithubOAuth struct {
	Access_token string `json:"access_token"`
	Username *github.User `json:"user"`
}

var (
	cfg      *Config
	oauthCfg *oauth2.Config
	store    *sessions.CookieStore

	// scopes
	scopes = []string{"repo"}
)

func loadConfig(file string) (*Config, error) {
	var config Config

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(b, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func saveOAuth(oauth GithubOAuth) {
	fo, err := os.Create(defaultOAuthFile)
	if err != nil {
		panic(err)
	}
	defer fo.Close()
	e := json.NewEncoder(fo)
	if err := e.Encode(oauth); err != nil {
		panic(err)
	}
}

func refreshAccessToken(token string)  {
	tc := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
	client := github.NewClient(tc)
	user, _, err := client.Users.Get(context.Background(), "")
	if err == nil {
		saveOAuth(GithubOAuth{
			Access_token: token,
			Username: user,
		})
	}
}

func GetGithubAccessToken() *GithubOAuth {
	var oAuth GithubOAuth

	b, err := ioutil.ReadFile(defaultOAuthFile)
	if err != nil {
		return nil
	}

	if err = json.Unmarshal(b, &oAuth); err != nil {
		return nil
	}

	return &oAuth
}

func StartGithubHandlerServer() *mux.Router {
	var err error
	cfg, err = loadConfig(defaultConfigFile)
	if err != nil {
		panic(err)
	}
	store = sessions.NewCookieStore([]byte(cfg.Secret))
	oauthCfg = &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  githubAuthorizeUrl,
			TokenURL: githubTokenUrl,
		},
		RedirectURL: redirectUrl,
		Scopes:      scopes,
	}

	r := mux.NewRouter()
	r.HandleFunc("/github", HomeHandler)
	r.HandleFunc("/github_auth", StartHandler)
	r.HandleFunc("/github_callback", CallbackHandler)

	//http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	//http.Handle("/", r)
	return r
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, SESSION)
	if err != nil {
		fmt.Fprintln(w, "aborted")
		return
	}
	token := session.Values[ACCESS_TOKEN]
	user := session.Values[USERNAME]
	var body string
	if token != nil {
		body = fmt.Sprintf("<h3>You are logged in</h1>\n <p>AccessToken: %s</p> <p>Username: %s</p>", token,
			user)
		refreshAccessToken(token.(string))
	} else {
		body = "<a href=\"/github_auth\">Get Github token </a>"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, body)
}

func StartHandler(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	session, _ := store.Get(r, SESSION)
	session.Values[STATE] = state
	sessions.Save(r, w)
	url := oauthCfg.AuthCodeURL(state)
	http.Redirect(w, r, url, 302)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, SESSION)
	if err != nil {
		fmt.Fprintln(w, "aborted")
		return
	}
	if r.URL.Query().Get("state") != session.Values[STATE] {
		fmt.Fprintln(w, "no state match; possible csrf OR cookies not enabled")
		return
	}
	token, err := oauthCfg.Exchange(context.Background(), r.URL.Query().Get("code"))
	if err != nil {
		fmt.Fprintln(w, "there was an issue getting your token")
		return
	}
	if !token.Valid() {
		fmt.Fprintln(w, "retreived invalid token")
		return
	}
	client := github.NewClient(oauthCfg.Client(context.Background(), token))
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		fmt.Println(w, "error getting name")
		return
	}
	session.Values[USERNAME] = user.Name
	session.Values[ACCESS_TOKEN] = token.AccessToken
	session.Save(r, w)
	saveOAuth(GithubOAuth{
		Access_token: token.AccessToken,
		Username: user,
	})
	http.Redirect(w, r, "/github", 302)
}
