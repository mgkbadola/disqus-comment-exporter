package main

import (
	"errors"
	"github.com/jessevdk/go-flags"
	"github.com/mgkbadola/disqus-comment-exporter/cmd"
	"os"
)

type Opts struct {
	ExportCmd cmd.ExportCommand `command:"export"`
}

func main() {
	var opts Opts
	p := flags.NewParser(&opts, flags.Default)
	p.CommandHandler = func(command flags.Commander, args []string) error {
		c := command.(cmd.CommonOptionsCommander)
		err := c.Execute(args)
		if err != nil {
			return err
		}
		return nil
	}
	if _, err := p.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}
}
