package cmd

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/euclid1990/gstats/configs"
	"github.com/euclid1990/gstats/utilities"
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
	googleOauth  *utilities.GoogleOauth
	googleClient *http.Client
)

// Action defines the main action for gstats
func Action(c *cli.Context) {
	exec := c.String("exec")
	fmt.Printf("Action: %v\n", exec)

	switch exec {
	case configs.ACTION_ALL:
	case configs.ACTION_INIT:
		googleOauth = utilities.NewGoogleOauth()
		go utilities.Server(googleOauth)
		googleClient = utilities.CreateGoogleClient(googleOauth)
		fmt.Printf("Google Client: %v\n", googleClient)
	}
}
