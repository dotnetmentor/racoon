package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/joho/godotenv"
)

var cases = []struct {
	name       string
	manifest   string
	parameters []string
	output     string
}{
	{"context_local", "racoon.yaml", []string{"context=local"}, "dotenv"},
	{"context_dev", "racoon.yaml", []string{"context=dev"}, "dotenv"},
	{"context_prod", "racoon.yaml", []string{"context=prod"}, "dotenv"},
	{"context_local_tenant_demo1", "racoon.yaml", []string{"context=local", "tenant=demo1"}, "dotenv"},
	{"context_dev_tenant_demo1", "racoon.yaml", []string{"context=dev", "tenant=demo1"}, "dotenv"},
	{"context_prod_tenant_demo1", "racoon.yaml", []string{"context=prod", "tenant=demo1"}, "dotenv"},
}

func TestExportCommand(t *testing.T) {
	cleanupTestOutput()

	app, _ := createApp()
	for _, tt := range cases {
		t.Run(tt.manifest, func(t *testing.T) {
			resetEnvFile := "./testdata/reset.env"
			godotenv.Overload(resetEnvFile)
			envFile := fmt.Sprintf("./testdata/%s.env", tt.name)
			godotenv.Overload(envFile)
			expectedFile := fmt.Sprintf("./testdata/%s.expected", tt.name)
			actualFile := fmt.Sprintf("./testdata/%s.actual", tt.name)

			args := os.Args[0:1]
			args = append(args, fmt.Sprintf("-manifest=./testdata/%s", tt.manifest))
			args = append(args, "-loglevel=debug")

			for _, p := range tt.parameters {
				args = append(args, fmt.Sprintf("-parameter=%s", p))
			}
			args = append(args, "export")
			args = append(args, fmt.Sprintf("-output=%s", tt.output))
			args = append(args, fmt.Sprintf("-path=%s", actualFile))

			err := app.Run(args)
			if err != nil {
				t.Error(err)
			}

			expected, _ := os.ReadFile(expectedFile)
			actual, _ := os.ReadFile(actualFile)
			if string(expected) != string(actual) {
				t.Errorf("\ncase: %s\ncommand:%s\n\n%s", tt.name, args[1:], diff.LineDiff(string(expected), string(actual)))
			}
		})
	}
}

func cleanupTestOutput() {
	files, err := filepath.Glob("./testdata/*.actual")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			panic(err)
		}
	}
}
