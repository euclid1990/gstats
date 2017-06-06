package cmd

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/euclid1990/gstats/configs"
	"github.com/euclid1990/gstats/utilities"
	"log"
	"net/http"
)

// List of options
var Flags = []cli.Flag{
	cli.StringFlag{
		Name:  "exec",
		Value: "all",
		Usage: "Execute action you want to do",
	},
	cli.StringFlag{
		Name:  "file",
		Value: "all",
		Usage: "Setup file secret",
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

	file := c.String("file")
	switch exec {
	case configs.ACTION_ALL:
	case configs.ACTION_LOC:
		googleOauth = utilities.NewGoogleOauth()
		googleClient := utilities.CreateGoogleClient(googleOauth)
		spreadSheet := utilities.NewSheet(googleClient)
		err := spreadSheet.UpdateLocSpreadsheets()
		if err != nil {
			log.Fatal(err)
		}
	case configs.ACTION_INIT:
		googleOauth = utilities.NewGoogleOauth()
		githubOauth = utilities.NewGithubOauth()
		go utilities.Server(googleOauth, githubOauth)
		client = utilities.CreateGoogleClient(googleOauth)
		fmt.Printf("Google Client: %v\n", client)
		client = utilities.CreateGithubClient(githubOauth)
		fmt.Printf("Github Client: %v\n", client)
	case configs.ACTION_REDMINE:
		redmine := utilities.NewRedmine()
		fmt.Printf("%v\n", redmine.NotifyInprogressIssuesToChatwork())
	case configs.ACTION_CHATWORK:
		setup := utilities.Setup{}
		setup.SetupNotice()
	case configs.ACTION_SETUP:
		utilities.SurveyRun(file)
	}
}
