package main

import (
	"fmt"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/yukimemi/spyrun"
)

func execute(c *cli.Context) {
	var err error
	var tomlpath string

	if c.String("input") != "" {
		tomlpath = c.String("input")
	} else {
		tomlpath = "./spy.toml"
	}
	log.Println("tomlpath:", tomlpath)

	err = spyrun.Run(tomlpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run spyrun ! %s", err.Error())
		os.Exit(1)
	}
}

func main() { // {{{

	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.Author = "yukimemi"
	app.Email = "yukimemi@gmail.com"
	app.Usage = "Watch files and Execute command."

	app.Flags = GlobalFlags
	app.Commands = Commands
	app.CommandNotFound = CommandNotFound

	app.Action = execute

	app.Run(os.Args)
} // }}}
