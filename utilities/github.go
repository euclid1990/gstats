package utilities

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/euclid1990/gstats/configs"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	GITHUB_AUTHORIZE_URL = "https://github.com/login/oauth/authorize"
	GITHUB_TOKEN_URL     = "https://github.com/login/oauth/access_token"
	REDIRECT_URL         = ""
)

type Config struct {
	ClientSecret string `json:"client_secret"`
	ClientId     string `json:"client_id"`
}

type GithubOauth struct {
	config   *oauth2.Config
	codeChan chan string
}

func NewGithubOauth() *GithubOauth {
	return &GithubOauth{
		codeChan: make(chan string),
	}
}

func CreateGithubClient(g *GithubOauth) *http.Client {
	g.readConfig()
	client := g.getClient(context.Background())
	return client
}

func (g *GithubOauth) loadConfig(file string) (*Config, error) {
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

func (g *GithubOauth) readConfig() {
	cfg, err := g.loadConfig(configs.PATH_GITHUB_SECRET)
	if err != nil {
		log.Fatalf("[Github Oauth] Unable to read client secret file: %v", err)
		panic(err)
	}

	g.config = &oauth2.Config{
		ClientID:     cfg.ClientId,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  GITHUB_AUTHORIZE_URL,
			TokenURL: GITHUB_TOKEN_URL,
		},
		RedirectURL: REDIRECT_URL,
		Scopes:      strings.Split(configs.SCOPE_GITHUB, ","),
	}
}

func (g *GithubOauth) getClient(ctx context.Context) *http.Client {
	cacheFile := g.tokenCacheFile()
	token, err := g.tokenFromFile(cacheFile)
	if err != nil || len(token.AccessToken) == 0 {
		token = g.getTokenFromWeb()
		fmt.Printf("%v\n", token)
		g.saveToken(cacheFile, token)
	}
	return g.config.Client(ctx, token)
}

func (g *GithubOauth) tokenCacheFile() string {
	_, err := os.Stat(configs.PATH_GITHUB_OAUTH)
	if err != nil {
		f, err := os.Create(configs.PATH_GITHUB_OAUTH)
		if err != nil {
			log.Fatalf("[Github Oauth] Unable to create path to cached credential file. %v", err)
		}
		defer f.Close()
	}
	return configs.PATH_GITHUB_OAUTH
}

func (g *GithubOauth) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	defer f.Close()
	return token, err
}

func (g *GithubOauth) getTokenFromWeb() *oauth2.Token {
	authURL := g.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("[Github Oauth] Go to the following link: \n%v\n", authURL)
	code := <-g.codeChan
	token, err := g.config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("[Github Oauth] Unable to retrieve Github token from web %v", err)
	}
	return token
}

func (g *GithubOauth) saveToken(file string, token *oauth2.Token) {
	log.Printf("[Github Oauth] Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("[Github Oauth] Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func GetAdditionsPullRequestGitHub(loc *Loc) error {
	githubOauth := NewGithubOauth()
	client := CreateGithubClient(githubOauth)
	githubClient := github.NewClient(client)
	eg := errgroup.Group{}
	for key, pr := range loc.Pr {
		userName, repoName, pullNumber, err := ExtractPullRequestInfo(pr.Link)
		if err != nil {
			continue
		}
		numberKey := key
		eg.Go(func() error {
			pull, _, err := githubClient.PullRequests.Get(context.Background(), userName, repoName, pullNumber)
			if err != nil {
				return err
			}
			loc.Pr[numberKey].Loc = pull.GetAdditions()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}
