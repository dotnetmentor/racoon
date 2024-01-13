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
	"strings"

	"github.com/dotnetmentor/racoon/internal/config"
	api "github.com/dotnetmentor/racoon/internal/httpapi"
	"github.com/dotnetmentor/racoon/internal/io"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func UI(metadata config.AppMetadata, fs embed.FS) *cli.Command {
	return &cli.Command{
		Name:      "ui",
		Usage:     "Exposes a UI over HTTP",
		UsageText: "",
		Hidden:    false,
		Flags:     []cli.Flag{},
		Action: func(c *cli.Context) error {
			ctx, err := newContext(c, metadata, false)
			if err != nil {
				return err
			}

			ctx.Log.Infof("starting UI server")

			s := api.NewServer(api.Config{
				Log:       ctx.Log.WithFields(logrus.Fields{"component": "server"}),
				BasicAuth: configureBasicAuth(),
			})

			s.Router.Handle("/*", indexHandler(
				ctx,
				fs,
			))
			s.Router.Post("/api/compare", compareHandler(
				ctx,
				ctx.Manifest.Filepath(),
			))

			err = s.RunAndBlock()
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func configureBasicAuth() *api.BasicAuth {
	var auth *api.BasicAuth
	baUsername := os.Getenv("CENTRY_SERVE_USERNAME")
	baPassword := os.Getenv("CENTRY_SERVE_PASSWORD")

	if baUsername != "" && baPassword != "" {
		auth = &api.BasicAuth{
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

func compareHandler(ctx config.AppContext, manifestPath string) func(w http.ResponseWriter, r *http.Request) {
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
		response := api.CompareResponse{}

		var body api.CompareRequest

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
			response.Left = &api.ExecutionResult{
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
			response.Right = &api.ExecutionResult{
				Logs:   stderr,
				Result: stdout,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		js, err := json.Marshal(response)
		if err == nil {
			w.Write(js)
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
