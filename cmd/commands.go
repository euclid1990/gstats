package cmd

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/euclid1990/gstats/utilities"
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
}
