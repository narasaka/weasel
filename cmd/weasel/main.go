package main

import (
	"context"
	"log"
	"os"

	"github.com/narasaka/weasel/helpers"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "weasel",
		Usage: "check for broken links",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"r"},
				Usage:   "recursively check links within the same domain",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() == 0 {
				return cli.ShowAppHelp(c)
			}
			helpers.Check(c.Args().First(), c.Bool("recursive"))
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
