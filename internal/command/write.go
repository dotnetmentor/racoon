package command

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ttacon/chalk"

	"github.com/dotnetmentor/racoon/internal/api"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/dotnetmentor/racoon/internal/visitor"

	"github.com/urfave/cli/v2"
)

func Write(metadata config.AppMetadata) *cli.Command {
	return &cli.Command{
		Name:  "write",
		Usage: "write values for properties defined in the manifest file",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "parameter",
				Aliases: []string{"p"},
				Usage:   "sets layer parameters",
			},
		},
		Action: func(c *cli.Context) error {
			ctx, err := newContext(c, metadata, true)
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

			type writable struct {
				layer        api.Layer
				value        api.Value
				formatter    api.ValueFormatter
				selectPrompt string
			}

			type valueInfo struct {
				layer       string
				property    string
				description string
				sensitive   bool
				source      string
				sourceKey   string
				formatter   string
			}

			promptForTargets := func(values []writable, msg string) (selected []writable) {
				options := make([]string, len(values))
				for i, v := range values {
					options[i] = v.selectPrompt
				}
				selection := make([]string, 0)
				prompt := &survey.MultiSelect{
					Message: msg + ":",
					Options: options,
				}
				survey.AskOne(prompt, &selection)
				for _, s := range selection {
					for _, v := range values {
						if v.selectPrompt == s {
							selected = append(selected, v)
						}
					}
				}
				return
			}

			promptForPropertyValue := func(msg string) string {
				fmt.Printf("%s? %s%s: ", chalk.Green, chalk.White, msg)
				reader := bufio.NewReader(os.Stdin)
				value, _ := reader.ReadString('\n')
				value = strings.TrimSuffix(value, "\n")
				return value
			}

			promptYesNo := func(msg string) bool {
				fmt.Printf("%s? %s%s (yes/no): ", chalk.Green, chalk.White, msg)
				reader := bufio.NewReader(os.Stdin)
				value, _ := reader.ReadString('\n')
				value = strings.TrimSuffix(value, "\n")
				if value == "yes" || value == "y" {
					return true
				}
				return false
			}

			visit := visitor.New(ctx)

			setNewValue := func(i valueInfo, p api.Property, v api.Value, sourceConfig config.SourceConfig, ctx config.AppContext) (api.Value, error) {
				infoFmt := "%s# %s%s%s: %s\n"
				fmt.Println()
				fmt.Printf(infoFmt, chalk.Magenta, chalk.Cyan, "layer", chalk.White, i.layer)
				fmt.Printf(infoFmt, chalk.Magenta, chalk.Cyan, "property", chalk.White, i.property)
				fmt.Printf(infoFmt, chalk.Magenta, chalk.Cyan, "desription", chalk.White, i.description)
				fmt.Printf(infoFmt, chalk.Magenta, chalk.Cyan, "sensitive", chalk.White, fmt.Sprintf("%v", i.sensitive))
				fmt.Printf(infoFmt, chalk.Magenta, chalk.Cyan, "source", chalk.White, i.source)
				fmt.Printf(infoFmt, chalk.Magenta, chalk.Cyan, "source key", chalk.White, i.sourceKey)
				if len(i.formatter) > 0 {
					fmt.Printf(infoFmt, chalk.Magenta, chalk.Yellow, "formatter", chalk.White, i.formatter)
				}

				if v.Error() == nil && i.sensitive && promptYesNo("preview current value") {
					// NOTE: This is one of the few time where it is OK to print the raw value
					fmt.Printf("%s! %scurrent value:%s %s\n", chalk.Green, chalk.White, chalk.Blue, v.Raw())
				} else {
					fmt.Printf("%s! %scurrent value:%s %s\n", chalk.Green, chalk.White, chalk.Blue, v.String())
				}

				for {
					strVal := promptForPropertyValue("new value")
					newVal := api.NewValue(v.Source(), i.sourceKey, strVal, nil, i.sensitive)

					if len(i.formatter) > 0 {
						// TODO: Validation for formatter source?
					} else {
						if err := p.Validate(newVal); err != nil {
							fmt.Printf("%s! %serror: %v\n", chalk.Red, chalk.White, err)
							continue
						}
					}

					if strVal == v.Raw() {
						break
					}

					ctx.Log.Debugf("setting %s in %s, new value: %s", i.sourceKey, v.Source().Type(), newVal.String())
					err := visit.Store().Write(i.sourceKey, strVal, i.description, v.Source().Type(), sourceConfig)
					if err != nil {
						return nil, err
					}
					return newVal, nil

				}
				return nil, nil
			}

			err = visit.Init(excludes, includes)
			if err != nil {
				return err
			}

			if err = visit.Property(func(p api.Property, err error) (bool, error) {
				if err != nil {
					return false, err
				}

				val := p.Value()

				// NOTE: We should not return error for invalid value at this point, it will stop us writing the initial value
				if err := p.Validate(val); err != nil {
					if val != nil {
						ctx.Log.Warnf("property %s, defined in %s, resolved to invalid value from %s, value: %s", p.Name, p.Source(), val.Source(), val.String())
					}
				} else {
					if val != nil {
						ctx.Log.Infof("property %s, defined in %s, resolved to value from %s, value: %s", p.Name, p.Source(), val.Source(), val.String())
					}
				}

				for _, v := range p.Values() {
					if err := p.Validate(v); err != nil {
						ctx.Log.Infof("- value from %s is invalid, err: %v", v.Source(), err)
					} else {
						ctx.Log.Infof("- value from %s, value: %s", v.Source(), v.String())
					}
				}

				wSources := make([]writable, 0)
				for _, v := range p.Values().Writable() {
					wSources = append(wSources, writable{
						layer:        v.Source().Layer(),
						value:        v,
						selectPrompt: v.SourceAndKey(),
					})
				}

				wFormatters := make([]writable, 0)
				if err := visit.Layer(func(l api.Layer, err error) (bool, error) {
					if lp := l.Property(p.Name); lp != nil {
						for _, fc := range lp.WritableFormatters() {
							f := api.NewFormatter(fc, ctx.Log)
							val := visit.Store().Read(l, f.FormattingKey(), p.Sensitive() || lp.Sensitive(), f.Source(), l.Config)
							wFormatters = append(wFormatters, writable{
								layer:        l,
								value:        val,
								formatter:    f,
								selectPrompt: fmt.Sprintf("%s, %s", f.String(), val.SourceAndKey()),
							})
						}
					}
					return true, nil
				}); err != nil {
					return false, err
				}

				ok := false
				if len(wSources) > 0 || len(wFormatters) > 0 {
					fmt.Println()
					ok = promptYesNo(fmt.Sprintf("set new value(s) for %s", p.Name))
					if !ok {
						fmt.Println()
						return true, nil
					}
				}

				if len(wSources) > 0 && ok {
					fmt.Println()
					for _, target := range promptForTargets(wSources, "property sources, select target(s) to update") {
						i := valueInfo{
							layer:       target.layer.Name,
							property:    p.Name,
							description: p.Description,
							sensitive:   p.Sensitive() || target.value.Sensitive(),
							source:      string(target.value.Source().Type()),
							sourceKey:   target.value.Key(),
						}
						nval, err := setNewValue(i, p, target.value, target.layer.Config, ctx)
						if err != nil {
							return false, err
						}

						if nval != nil {
							p.SetValue(nval)
						}
					}
				}

				if len(wFormatters) > 0 && ok {
					fmt.Println()
					selected := promptForTargets(wFormatters, "the property uses formatters to construct it's final value, select target(s) to update")
					for _, target := range selected {
						lp := target.layer.Property(p.Name)
						i := valueInfo{
							layer:       target.layer.Name,
							property:    p.Name,
							description: p.Description,
							sensitive:   p.Sensitive() || lp.Sensitive() || target.value.Sensitive(),
							source:      string(target.value.Source().Type()),
							sourceKey:   target.value.Key(),
							formatter:   target.formatter.String(),
						}
						nval, err := setNewValue(i, *lp, target.value, target.layer.Config, ctx)
						if err != nil {
							return false, err
						}

						if nval != nil {
							lp.SetValue(nval)
						}
					}
				}

				// TODO: Fix so that at the end of visiting a property, the value is up to date
				//ctx.Log.Warnf("property %s, value after processing: %s", p.Name, p.Value().String())

				return true, nil
			}); err != nil {
				return err
			}

			return nil
		},
	}
}
