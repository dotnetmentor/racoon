package command

import (
	"fmt"
	"strings"

	"github.com/dotnetmentor/racoon/internal/api"
	"github.com/dotnetmentor/racoon/internal/visitor"
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

			visit := visitor.New(ctx)

			err = visit.Init([]string{}, []string{key})
			if err != nil {
				return err
			}

			var value api.Value
			err = visit.Property(func(p api.Property, err error) error {
				if err != nil {
					return err
				}

				val := p.Value()
				if val == nil {
					return fmt.Errorf("no value resolved for property %s", p.Name)
				}

				if val.Error() != nil {
					return fmt.Errorf("no value resolved for property %s, err: %w", p.Name, val.Error())
				}

				if err := p.Validate(val); err != nil {
					return err
				}

				ctx.Log.Debugf("property %s, defined in %s, value from %s, value set to: %s", p.Name, p.Source(), val.Source(), val.String())

				value = val
				return nil
			})
			if err != nil {
				return err
			}

			fmt.Printf("%s", value.Raw())

			return nil
		},
	}
}
