package utilities

import (
	"encoding/json"
	"fmt"
	"github.com/euclid1990/gstats/configs"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type GoogleOauth struct {
	config   *oauth2.Config
	codeChan chan string
}

func NewGoogleOauth() *GoogleOauth {
	return &GoogleOauth{codeChan: make(chan string)}
}

func CreateGoogleClient(g *GoogleOauth) *http.Client {
	ctx := context.Background()
	g.readConfig()
	client := g.getClient(ctx)
	return client
}

func (g *GoogleOauth) readConfig() {
	b, err := ioutil.ReadFile(configs.PATH_GOOGLE_SECRET)
	if err != nil {
		log.Fatalf("[Google Oauth] Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, configs.SCOPE_GOOGLE_SPREADSHEET)
	if err != nil {
		log.Fatalf("[Google Oauth] Unable to parse client secret file to config: %v", err)
	}
	g.config = config
}

func (g *GoogleOauth) getClient(ctx context.Context) *http.Client {
	cacheFile := g.tokenCacheFile()
	tok, err := g.tokenFromFile(cacheFile)
	if err != nil {
		tok = g.getTokenFromWeb()
		g.saveToken(cacheFile, tok)
	}
	// Renew expired token
	if !tok.Valid() {
		src := g.config.TokenSource(ctx, tok)
		newToken, err := src.Token()
		if err != nil {
			log.Fatalf("[Google Oauth] Could not renew a token using a RefreshToken. %v", err)
		}
		g.saveToken(cacheFile, newToken)
		return g.config.Client(ctx, newToken)
	}
	return g.config.Client(ctx, tok)
}

func (g *GoogleOauth) tokenCacheFile() string {
	_, err := os.Stat(configs.PATH_GOOGLE_OAUTH)
	if err != nil {
		file, err := os.Create(configs.PATH_GOOGLE_OAUTH)
		defer file.Close()
		if err != nil {
			log.Fatalf("[Google Oauth] Unable to create path to cached credential file. %v", err)
		}
	}
	return configs.PATH_GOOGLE_OAUTH
}

func (g *GoogleOauth) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

func (g *GoogleOauth) getTokenFromWeb() *oauth2.Token {
	// To get the refresh token, add approval_prompt=force
	authURL := g.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("- [Google Oauth] Go to the following link: \n%v\n", authURL)
	code := <-g.codeChan
	tok, err := g.config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("[Google Oauth] Unable to retrieve Google token from web %v", err)
	}
	return tok
}

func (g *GoogleOauth) saveToken(file string, token *oauth2.Token) {
	log.Printf("[Google Oauth] Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("[Google Oauth] Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
