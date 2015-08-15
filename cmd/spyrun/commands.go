package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

var GlobalFlags = []cli.Flag{
	cli.StringFlag{
		// EnvVar: "ENV_I",
		Name: "input, i",
		// Value: "",
		Usage: "Input config file",
	},
}

var Commands = []cli.Command{}

func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
}
