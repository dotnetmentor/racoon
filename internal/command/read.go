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
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "parameter",
				Aliases: []string{"p"},
				Usage:   "sets layer parameters",
			},
		},
		Action: func(c *cli.Context) error {
			if c.Args().Len() != 1 {
				return fmt.Errorf("key not specified, must be provided as a single argument")
			}
			key := strings.TrimSpace(c.Args().First())

			ctx, err := newContext(c, true)
			if err != nil {
				return err
			}

			visit := visitor.New(ctx)

			err = visit.Init([]string{}, []string{key})
			if err != nil {
				return err
			}

			var value api.Value
			err = visit.Property(func(p api.Property, err error) (bool, error) {
				if err != nil {
					return false, err
				}

				val := p.Value()
				if err := p.Validate(val); err != nil {
					return false, err
				}

				// If validation passes but the value is nil, continue
				if val == nil {
					return true, nil
				}

				// If validation passes but we have a not found error for the resolved value, skip read
				if !api.IsNotFoundError(val.Error()) {
					value = val
				}

				ctx.Log.Debugf("property %s, defined in %s, value from %s, value set to: %s", p.Name, p.Source(), val.Source(), val.String())
				for _, v := range p.Values() {
					if err := p.Validate(v); err != nil {
						ctx.Log.Debugf("- value from %s is invalid, err: %v", v.Source(), err)
					} else {
						ctx.Log.Debugf("- value from %s, value: %s", v.Source(), v.String())
					}
				}

				return true, nil
			})
			if err != nil {
				return err
			}

			if value != nil {
				fmt.Printf("%s", value.Raw())
			}

			return nil
		},
	}
}
