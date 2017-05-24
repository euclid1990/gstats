package main

import (
	"github.com/codegangsta/cli"
	"github.com/euclid1990/gstats/cmd"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "gstats"
	app.Version = "1.0.0"
	app.Usage = "A small cli written in Go to help update lines of code."
	app.Action = cmd.Action
	app.Flags = cmd.Flags
	app.Run(os.Args)
}
