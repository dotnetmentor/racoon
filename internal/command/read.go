package command

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

func Read() *cli.Command {
	return &cli.Command{
		Name:  "read",
		Usage: "reads a single value",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {
			if c.Args().Len() != 1 {
				return fmt.Errorf("key not specified, must be provided as a single argument")
			}
			key := strings.TrimSpace(c.Args().First())

			ctx, err := getContext(c)
			if err != nil {
				return err
			}

			vs := ValueSource{
				context:    ctx,
				properties: make([]Property, 0),
			}

			value, err := vs.ReadOne(key)
			if err != nil {
				return err
			}

			fmt.Printf("%s", value.Raw())
			return nil
		},
	}
}
