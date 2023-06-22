package command

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/dotnetmentor/racoon/internal/output"
	"github.com/dotnetmentor/racoon/internal/utils"

	"github.com/urfave/cli/v2"
)

func Export() *cli.Command {
	return &cli.Command{
		Name:  "export",
		Usage: "export values",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "export using output",
			},
			&cli.StringFlag{
				Name:    "alias",
				Aliases: []string{"a"},
				Usage:   "export using output matching alias",
			},
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "export output to the specified path",
			},
			&cli.StringSliceFlag{
				Name:    "include",
				Aliases: []string{"i"},
				Usage:   "include property in export",
			},
			&cli.StringSliceFlag{
				Name:    "exclude",
				Aliases: []string{"e"},
				Usage:   "exclude property from export",
			},
		},
		Action: func(c *cli.Context) error {
			ctx, err := getContext(c)
			if err != nil {
				return err
			}
			m := ctx.Manifest

			ot := c.String("output")
			oa := c.String("alias")
			p := c.String("path")

			includes := c.StringSlice("include")
			excludes := c.StringSlice("exclude")

			if ot == "" && p != "" {
				ctx.Log.Warn("the flag --path is not allowed without also specifying the --output flag")
				return nil
			}

			vs := &ValueSource{
				context:    ctx,
				properties: make([]Property, 0),
			}

			keys, values, err := vs.ReadAll(excludes, includes)
			if err != nil {
				return err
			}

			// output
			outputMatched := false
			for _, o := range m.Outputs {
				if ot != "" && string(o.Type) != ot {
					continue
				}

				if oa != "" && o.Alias != oa {
					oid := string(o.Type)
					if len(o.Alias) > 0 {
						oid = fmt.Sprintf("%s/%s", oid, o.Alias)
					}
					ctx.Log.Debugf("skipping %s output, did not match the alias %s", oid, oa)
					continue
				}

				outputMatched = true

				path := o.Path
				if p != "" {
					path = p
				}

				if ot == "" && path == "-" {
					ctx.Log.Infof("writing to stdout is only allowed when using the --output flag, skipping output %s", o.Type)
					continue
				}

				filtered := []string{}
				filteredValues := make(map[string]string)
				for _, s := range keys {
					if len(o.Exclude) > 0 && utils.StringSliceContains(o.Exclude, s) {
						continue
					}
					if len(o.Include) > 0 && !utils.StringSliceContains(o.Include, s) {
						continue
					}

					switch o.Export {
					case config.ExportTypeClearText:
						switch values[s].(type) {
						case *SensitiveValue:
							continue
						}
					case config.ExportTypeSensitive:
						switch values[s].(type) {
						case *ClearTextValue:
							continue
						}
					}

					filtered = append(filtered, s)
					filteredValues[s] = values[s].Raw()
				}

				out := os.Stdout
				if path != "" && path != "-" {
					file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
					if err != nil {
						return fmt.Errorf("failed to open file for writing, %v", err)
					}
					defer file.Close()
					defer file.Sync()
					out = file
				}
				w := bufio.NewWriter(out)
				defer w.Flush()

				switch out := config.AsOutput(o).(type) {
				case output.Dotenv:
					ctx.Log.Infof("exporting values as dotenv ( path=%s quote=%v )", path, out.Quote)
					out.Write(w, filtered, o.Map, filteredValues)
				case output.Tfvars:
					ctx.Log.Infof("exporting values as tfvars ( path=%s )", path)
					out.Write(w, filtered, o.Map, filteredValues)
				case output.Json:
					ctx.Log.Infof("exporting values as json ( path=%s )", path)
					out.Write(w, filtered, o.Map, filteredValues)
				default:
					return fmt.Errorf("unsupported output type %s", o.Type)
				}
			}

			if ot != "" && !outputMatched {
				return fmt.Errorf("unknown output type %s", ot)
			}

			return nil
		},
	}
}
