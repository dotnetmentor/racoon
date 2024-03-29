package command

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dotnetmentor/racoon/internal/api"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/dotnetmentor/racoon/internal/output"
	"github.com/dotnetmentor/racoon/internal/utils"
	"github.com/dotnetmentor/racoon/internal/visitor"

	"github.com/urfave/cli/v2"
)

func Export(metadata config.AppMetadata) *cli.Command {
	return &cli.Command{
		Name:  "export",
		Usage: "Exports the values of multiple properties",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "parameter",
				Aliases: []string{"p"},
				Usage:   "sets layer parameters",
			},
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
				Aliases: []string{}, // 'p' not possible as alias
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
			ctx, err := newContext(c, metadata, true)
			if err != nil {
				return err
			}
			m := ctx.Manifest

			ot := c.String("output")
			oa := c.String("alias")
			p := c.String("path")

			excludes := c.StringSlice("exclude")
			includes := c.StringSlice("include")

			if ot == "" && p != "" {
				ctx.Log.Warn("the flag --path is not allowed without also specifying the --output flag")
				return nil
			}

			keys := []string{}
			values := map[string]api.Value{}

			backend, err := newBackend(ctx)
			if err != nil {
				return err
			}

			encconf := api.NewEncryptedConfig(ctx, backend)

			visit := visitor.New(ctx)

			err = visit.Init(excludes, includes)
			if err != nil {
				return err
			}

			err = visit.Property(func(p api.Property, err error) (bool, error) {
				if err != nil {
					return false, err
				}

				if err := encconf.Track(p); err != nil {
					return false, err
				}

				key := p.Name

				if !utils.StringSliceContains(keys, key) {
					keys = append(keys, key)
				}

				val := p.Value()
				if err := p.Validate(val); err != nil {
					return false, err
				}

				// If validation passes but the value is nil, continue
				if val == nil {
					return true, nil
				}

				// If validation passes but we have a not found error for the resolved value, skip export
				if !api.IsNotFoundError(val.Error()) {
					values[key] = val
				}

				ctx.Log.Infof("property %s, defined in %s, value from %s, value set to: %s", p.Name, p.Source(), val.Source(), val.String())
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

			// track encrypted config
			if backend != nil {
				jb, err := json.Marshal(&encconf)
				if err != nil {
					return err
				}

				if err := backend.Store().Upload(encconf.Path(), jb); err != nil {
					return err
				}
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

				paths := o.Paths
				if p != "" {
					paths = []string{p}
				}

				for _, path := range paths {
					if ot == "" && path == "-" {
						ctx.Log.Infof("writing to stdout is only allowed when using the --output flag, skipping output %s (alias=%s path=%s)", o.Type, o.Alias, path)
						continue
					}

					path = ctx.Replace(path)

					filtered := []string{}
					filteredValues := make(map[string]string)
					for _, s := range keys {
						if len(o.Exclude) > 0 && utils.StringSliceContains(o.Exclude, s) {
							continue
						}
						if len(o.Include) > 0 && !utils.StringSliceContains(o.Include, s) {
							continue
						}

						v, ok := values[s]
						if !ok {
							continue
						}

						switch o.Export {
						case config.ExportTypeClearText:
							switch v.(type) {
							case *api.SensitiveValue:
								continue
							}
						case config.ExportTypeSensitive:
							switch v.(type) {
							case *api.ClearTextValue:
								continue
							}
						}

						filtered = append(filtered, s)
						filteredValues[s] = v.Raw()
					}

					err := func() error {
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
							ctx.Log.Infof("exporting values as dotenv (alias=%s path=%s quote=%v)", o.Alias, path, out.Quote)
							out.Write(w, filtered, o.Map, filteredValues)
						case output.Tfvars:
							ctx.Log.Infof("exporting values as tfvars (alias=%s path=%s)", o.Alias, path)
							out.Write(w, filtered, o.Map, filteredValues)
						case output.Json:
							ctx.Log.Infof("exporting values as json (alias=%s path=%s)", o.Alias, path)
							out.Write(w, filtered, o.Map, filteredValues)
						default:
							return fmt.Errorf("unsupported output type %s", o.Type)
						}

						return nil
					}()
					if err != nil {
						return err
					}
				}
			}

			if ot != "" && !outputMatched {
				return fmt.Errorf("unknown output (type=%s alias=%s)", ot, oa)
			}

			return nil
		},
	}
}
