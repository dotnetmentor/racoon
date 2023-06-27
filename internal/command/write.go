package command

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ttacon/chalk"

	"github.com/dotnetmentor/racoon/internal/api"
	"github.com/dotnetmentor/racoon/internal/visitor"

	"github.com/urfave/cli/v2"
)

func Write() *cli.Command {
	return &cli.Command{
		Name:  "write",
		Usage: "write values for properties defined in the manifest file",
		Action: func(c *cli.Context) error {
			ctx, err := getContext(c)
			if err != nil {
				return err
			}

			excludes := []string{}
			includes := []string{}

			if c.NArg() > 0 {
				for _, a := range c.Args().Slice() {
					key := strings.TrimSpace(a)
					includes = append(includes, key)
				}
			}

			promptForDestination := func(values []string) (reply string) {
				prompt := &survey.Select{
					Message: "select writable destination",
					Options: values,
				}
				survey.AskOne(prompt, &reply)
				return
			}

			promptForValue := func(p api.Property) string {
				fmt.Printf("%s? %s%s (%s): ", chalk.Green, chalk.White, p.Name, p.Description)
				reader := bufio.NewReader(os.Stdin)
				value, _ := reader.ReadString('\n')
				value = strings.TrimSuffix(value, "\n")
				return value
			}

			promptYesNo := func(msg string) bool {
				fmt.Printf("%s? %s%s (yes/no) ", chalk.Green, chalk.White, msg)
				reader := bufio.NewReader(os.Stdin)
				value, _ := reader.ReadString('\n')
				value = strings.TrimSuffix(value, "\n")
				if value == "yes" || value == "y" {
					return true
				}
				return false
			}

			visit := visitor.New(ctx)

			err = visit.Init(excludes, includes)
			if err != nil {
				return err
			}

			err = visit.Property(func(p api.Property, err error) error {
				if err != nil {
					return err
				}

				val := p.Value()
				if val == nil {
					return fmt.Errorf("no value resolved for property %s", p.Name)
				}

				if val.Error() != nil && !api.IsNotFoundError(val.Error()) {
					return val.Error()
				}

				if err := p.Validate(val); err != nil {
					ctx.Log.Warnf("property %s, defined in %s, resolved to invalid value from %s, value: %s", p.Name, p.Source(), val.Source(), val.String())
				} else {
					ctx.Log.Infof("property %s, defined in %s, resolved to value from %s, value: %s", p.Name, p.Source(), val.Source(), val.String())
				}
				for _, v := range p.Values() {
					if err := p.Validate(v); err != nil {
						ctx.Log.Infof("- value from %s is invalid, err: %v", v.Source(), err)
					} else {
						ctx.Log.Infof("- value from %s, value: %s", v.Source(), v.String())
					}
				}

				writable := p.Values().Writable()
				if len(writable) == 0 {
					ctx.Log.Debugf("no writable sources found for property %s, skipping write...", p.Name)
					return nil
				}
				if len(writable) > 0 && promptYesNo(fmt.Sprintf("set new value for %s", p.Name)) {
					destinations := make([]string, len(writable))
					for i, wd := range writable {
						destinations[i] = wd.SourceAndKey()
					}
					selected := promptForDestination(destinations)
					if len(selected) > 0 {
						var dest api.Value
						for _, v := range writable {
							if v.SourceAndKey() == selected {
								dest = v
							}
						}

						if !api.IsNotFoundError(dest.Error()) && dest.Sensitive() && promptYesNo("preview current value (sensitive)") {
							// NOTE: This is one of the few time where it is OK to print the raw value
							fmt.Printf("%s! %scurrent value:%s %s\n", chalk.Green, chalk.White, chalk.Cyan, dest.Raw())
						}

						for {
							strVal := promptForValue(p)
							newVal := api.NewValue(dest.Source(), dest.Key(), strVal, nil, dest.Sensitive())
							if err := p.Validate(newVal); err != nil {
								fmt.Printf("%s! %serror: %v\n", chalk.Red, chalk.White, err)
								continue
							}

							if strVal == dest.Raw() {
								break
							}

							ctx.Log.Infof("setting property %s in %s, new value: %s", p.Name, dest.Source(), newVal.String())
							err := visit.Store().Write(dest.Key(), strVal, p.Description, dest.Source().Type(), ctx.Manifest.Config.Sources)
							if err != nil {
								return err
							}
							p.SetValue(newVal)
							break
						}
					}
				}

				// TODO: Handle write for formatting sources (writable)

				return nil
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
}
