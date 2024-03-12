package command

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/dotnetmentor/racoon/internal/api"
	"github.com/dotnetmentor/racoon/internal/backend"
	"github.com/dotnetmentor/racoon/internal/config"
	"github.com/dotnetmentor/racoon/internal/httpapi"
	"github.com/dotnetmentor/racoon/internal/io"
	"github.com/dotnetmentor/racoon/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func UI(metadata config.AppMetadata, fs embed.FS) *cli.Command {
	return &cli.Command{
		Name:      "ui",
		Usage:     "Exposes the racoon UI over HTTP",
		UsageText: "",
		Hidden:    false,
		Flags:     []cli.Flag{},
		Action: func(c *cli.Context) error {
			ctx, err := newContext(c, metadata, false)
			if err != nil {
				return err
			}

			ctx.Log.Infof("starting UI server")

			s := httpapi.NewServer(httpapi.Config{
				Log:       ctx.Log.WithFields(logrus.Fields{"component": "server"}),
				BasicAuth: configureBasicAuth(),
			})

			s.Router.Handle("/*", indexHandler(
				ctx,
				fs,
			))
			s.Router.Post("/api/command/compare", compareCommandHandler(
				ctx,
				ctx.Manifest.Filepath(),
			))
			s.Router.Post("/api/command/config/decrypt", decryptConfigCommandHandler(ctx))
			s.Router.Get("/api/query/config", configQueryHandler(ctx))

			err = s.RunAndBlock()
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func configureBasicAuth() *httpapi.BasicAuth {
	var auth *httpapi.BasicAuth
	baUsername := os.Getenv("CENTRY_SERVE_USERNAME")
	baPassword := os.Getenv("CENTRY_SERVE_PASSWORD")

	if baUsername != "" && baPassword != "" {
		auth = &httpapi.BasicAuth{
			Username: baUsername,
			Password: baPassword,
		}
	}

	return auth
}

func indexHandler(ctx config.AppContext, embedded embed.FS) http.HandlerFunc {
	assetsFs, err := fs.Sub(embedded, "ui/dist")
	if err != nil {
		panic(err)
	}
	assetHandler := http.FileServer(http.FS(assetsFs))
	redirector := createRedirector(ctx, assetsFs)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extension := filepath.Ext(r.URL.Path)
		// We use the golang http.FileServer for static file requests.
		// This will return a 404 on normal page requests, ie /kustomizations and /sources.
		// Redirect all non-file requests to index.html, where the JS routing will take over.
		if extension == "" {
			redirector(w, r)
			return
		}
		assetHandler.ServeHTTP(w, r)
	})
}

func configQueryHandler(ctx config.AppContext) func(w http.ResponseWriter, r *http.Request) {
	backend, err := backend.New(ctx.Context, ctx.Manifest.Backend)
	if err != nil {
		ctx.Log.Fatal(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx.Log.Infof("querying configs")
		statusCode := http.StatusOK
		response := httpapi.ConfigQueryResponse{
			Items: make([]httpapi.ConfigQueryItem, 0),
		}

		if files, err := backend.Store().List(); err != nil {
			response.Error = fmt.Sprintf("error listing configs: %v", err)
			statusCode = http.StatusInternalServerError
		} else {
			download := false
			filters := r.URL.Query()["f"]

			if r.URL.Query().Get("download") == "true" {
				download = true
			}

			configs := make([]httpapi.ConfigQueryItem, 0)
			for _, file := range files {
				configs = append(configs, httpapi.ConfigQueryItem{
					Path:      file,
					Encrypted: true,
				})
			}
			if download {
				n := len(configs)
				configs = filterConfigs(configs, filters)
				ctx.Log.Infof("applying filters, yielded %d config(s), was %d", len(configs), n)
			}

			sort.Slice(configs, func(i, j int) bool {
				return configs[i].Matches > configs[j].Matches
			})

			response.Total = len(configs)
			for _, f := range filters {
				response.Filters = append(response.Filters, strings.ReplaceAll(f, "/", "="))
			}

			if download {
				startAt := 0
				if r.URL.Query().Get("startAt") != "" {
					pv, err := strconv.Atoi(r.URL.Query().Get("startAt"))
					if err == nil {
						startAt = pv
					}
				}
				configs = utils.SliceSkip(configs, startAt)
				ctx.Log.Infof("skip %d, yielded %d config(s)", startAt, len(configs))

				beforeTake := len(configs)
				take := 6 // TODO: Make this configurable
				configs = utils.SliceTake(configs, take)
				response.More = beforeTake > len(configs)
				ctx.Log.Infof("take %d, yielded %d config(s)", take, len(configs))

				for i, c := range configs {
					ctx.Log.Debugf("downloading config %s", c.Path)
					encrypted, err := backend.Store().Download(c.Path)
					if err != nil {
						response.Error = fmt.Sprintf("error downloading config %s: %v", c.Path, err)
						statusCode = http.StatusInternalServerError
						break
					}
					configs[i].Data = encrypted
					ctx.Log.Infof("downloaded config %s", c.Path)
				}
			}

			response.Items = configs
		}

		if response.Error != "" {
			ctx.Log.Error(response.Error)
		}

		if err := jsonRespone(w, statusCode, response); err != nil {
			ctx.Log.Errorf("error writing response: %v", err)
		}
	}
}

func filterConfigs(configs []httpapi.ConfigQueryItem, filters []string) (filtered []httpapi.ConfigQueryItem) {
	for _, c := range configs {
		matchForKey := make(map[string]bool)
		filtersByGroup := make(map[string][]string)
		for _, f := range filters {
			kv := strings.Split(f, "/")
			key := kv[0]
			value := kv[1]
			filtersByGroup[key] = append(filtersByGroup[key], value)
		}

		for key, values := range filtersByGroup {
			match := false
			for _, v := range values {
				searchStr := fmt.Sprintf("%s/%s", key, v)
				if key == "name" {
					searchStr = fmt.Sprintf("%s/%s", v, "racoon.config")
				}
				if strings.Contains(c.Path, searchStr) {
					match = true
					c.Matches++
				}
			}
			matchForKey[key] = match
		}

		matchesAllGroups := true
		for _, ok := range matchForKey {
			if !ok {
				matchesAllGroups = false
				break
			}
		}

		if matchesAllGroups {
			filtered = append(filtered, c)
		}
	}
	return
}

func decryptConfigCommandHandler(ctx config.AppContext) func(w http.ResponseWriter, r *http.Request) {
	backend, err := backend.New(ctx.Context, ctx.Manifest.Backend)
	if err != nil {
		ctx.Log.Fatal(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx.Log.Infof("decrypting config")
		statusCode := http.StatusOK
		response := httpapi.ConfigDecryptCommandResponse{}

		var body httpapi.ConfigDecryptCommand

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&body)
		if err != nil {
			ctx.Log.Errorf("error decoding request body: %v", err)
			statusCode = http.StatusBadRequest
		} else {
			ctx.Log.Debugf("downloading config %s", body.Path)
			encrypted, err := backend.Store().Download(body.Path)
			if err != nil {
				response.Error = fmt.Sprintf("error downloading config %s: %v", body.Path, err)
				statusCode = http.StatusInternalServerError
			} else {
				ctx.Log.Infof("downloaded config %s", body.Path)

				ctx.Log.Debugf("unmarshalling config %s", body.Path)
				encconf := api.EncryptedConfig{}
				if err := json.Unmarshal(encrypted, &encconf); err != nil {
					response.Error = fmt.Sprintf("error unmarshalling config %s: %v", body.Path, err)
					statusCode = http.StatusInternalServerError
				}

				for i, p := range encconf.Properties {
					if p.Sensitive && p.Value != nil && len(*p.Value) > 0 {
						ctx.Log.Infof("decrypting property %s", p.Name)
						ev := *p.Value
						dv, err := backend.Encryption().Decrypt([]byte(ev))
						if err != nil {
							response.Error = fmt.Sprintf("error decrypting property %s: %v", p.Name, err)
							statusCode = http.StatusInternalServerError
							ctx.Log.Error(response.Error)
							break
						}
						dsv := string(dv)
						encconf.Properties[i].Value = &dsv
					}
				}

				if response.Error == "" {
					ctx.Log.Debugf("marshalling config %s", body.Path)
					decrypted, err := json.Marshal(encconf)
					if err != nil {
						response.Error = fmt.Sprintf("error marshalling config %s: %v", body.Path, err)
						statusCode = http.StatusInternalServerError
						ctx.Log.Error(response.Error)
					} else {
						response.Data = decrypted
					}
				}
			}
		}

		if response.Error != "" {
			ctx.Log.Error(response.Error)
		}

		if err := jsonRespone(w, statusCode, response); err != nil {
			ctx.Log.Errorf("error writing response: %v", err)
		}
	}
}

func compareCommandHandler(ctx config.AppContext, manifestPath string) func(w http.ResponseWriter, r *http.Request) {
	runCommand := func(command, args string) (output string, logOutput string, err error) {
		io, stdout, stderr := io.Buffered(os.Stdin)
		execArgs := strings.Fields(fmt.Sprintf(
			"-m %s -l %s %s %s",
			manifestPath,
			ctx.Log.Level.String(),
			command,
			args,
		))

		ctx.Log.Infoln("executing command:", execArgs)

		self := os.Args[0]
		cmd := exec.Command(self, execArgs...)
		cmd.Stdin = io.Stdin
		cmd.Stdout = io.Stdout
		cmd.Stderr = io.Stderr

		// TODO: Can we trim sensitive data from stdout/stderr based on the declared inputs/outputs of a package

		if err := cmd.Run(); err != nil {
			return stdout.String(), stderr.String(), err
		}
		return stdout.String(), stderr.String(), nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx.Log.Infof("comparing command output")
		statusCode := http.StatusOK
		response := httpapi.CompareCommandResponse{}

		var body httpapi.CompareCommand

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&body)
		if err != nil {
			ctx.Log.Errorf("error decoding request body: %v", err)
			statusCode = http.StatusBadRequest
		}

		var stdout, stderr string

		// Left
		if body.Left != "" {
			stdout, stderr, err = runCommand(body.Command, body.Left)
			if err != nil {
				ctx.Log.Errorf("error executing compare left: %v", err)
			}
			response.Left = &httpapi.ExecutionResult{
				Logs:   stderr,
				Result: stdout,
			}
		}

		// Right
		if body.Right != "" {
			stdout, stderr, err = runCommand(body.Command, body.Right)
			if err != nil {
				ctx.Log.Errorf("error executing compare right: %v", err)
			}
			response.Right = &httpapi.ExecutionResult{
				Logs:   stderr,
				Result: stdout,
			}
		}

		if err := jsonRespone(w, statusCode, response); err != nil {
			ctx.Log.Errorf("error writing response: %v", err)
		}
	}
}

func createRedirector(ctx config.AppContext, fsys fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexPage, err := fsys.Open("index.html")

		if err != nil {
			ctx.Log.Debugf("could not open index.html page, %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		stat, err := indexPage.Stat()
		if err != nil {
			ctx.Log.Debugf("could not get index.html stat, %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bt := make([]byte, stat.Size())
		_, err = indexPage.Read(bt)

		if err != nil {
			ctx.Log.Debugf("could not read index.html, %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write(bt)

		if err != nil {
			ctx.Log.Debugf("error writing index.html, %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func jsonRespone(w http.ResponseWriter, statusCode int, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	js, err := json.Marshal(response)
	if err == nil {
		w.Write(js)
	}
	return nil
}
