package cmd

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/euclid1990/gstats/utilities"
	"log"
	"net/http"
	"os"
)

// List of options
var Flags = []cli.Flag{
	cli.StringFlag{
		Name:  "exec",
		Value: "all",
		Usage: "Execute action you want to do",
	},
}

// Action defines the main action for gstats
func Action(c *cli.Context) {
	exec := c.String("exec")
	fmt.Printf("Action: %v\n", exec)
	switch exec {
	case "auth":
		githubRoute := utilities.StartGithubHandlerServer()
		http.Handle("/", githubRoute)
		port := os.Getenv("PORT")
		if len(port) == 0 {
			port = "3000"
		}
		fmt.Printf("Action: %v\n", port)
		log.Fatalln(http.ListenAndServe(":" + port, nil))
		break
	case "github_token":
		token := utilities.GetGithubAccessToken()
		if token != nil {
			fmt.Printf("Github access token: %s\n", token.Access_token)
		} else {
			fmt.Print("Github access token: nil \n")
		}
	default:
		break
	}
}
