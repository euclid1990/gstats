package cmd

import (
	"context"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/euclid1990/gstats/configs"
	"github.com/euclid1990/gstats/utilities"
	"github.com/google/go-github/github"
	"net/http"
)

// List of options
var Flags = []cli.Flag{
	cli.StringFlag{
		Name:  "exec",
		Value: "all",
		Usage: "Execute action you want to do",
	},
}

// Instance of Google/Github Client
var (
	googleOauth *utilities.GoogleOauth
	githubOauth *utilities.GithubOauth
	client      *http.Client
)

// Action defines the main action for gstats
func Action(c *cli.Context) {
	exec := c.String("exec")
	fmt.Printf("Action: %v\n", exec)

	switch exec {
	case configs.ACTION_ALL:
	case configs.ACTION_INIT:
		googleOauth = utilities.NewGoogleOauth()
		githubOauth = utilities.NewGithubOauth()

		go utilities.Server(googleOauth, githubOauth)
		client = utilities.CreateGoogleClient(googleOauth)

		// Just for Testing
		fmt.Printf("Google Client: %v\n", client)
		utilities.NewSheet(client)

		client = utilities.CreateGithubClient(githubOauth)
		githubClient := github.NewClient(client)
		pulls, res, _ := githubClient.PullRequests.List(context.Background(), "euclid1990",
			"gstats", nil)
		fmt.Printf("API %v\n", res.Request.URL)
		for _, pull := range pulls {
			fmt.Printf("User: %s\nTitle: %s\nState: %s \n\n", pull.User.GetLogin(), pull.GetTitle(),
				pull.GetState())
		}
	case configs.ACTION_REDMINE:
		redmine := utilities.NewRedmine()
		idArray := []int{1288, 1315}
		fmt.Printf("%v\n", redmine.GetIds(idArray))
	}
}
