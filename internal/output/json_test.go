package output_test

import (
	"io"
	"os"

	pio "github.com/dotnetmentor/racoon/internal/io"
	"github.com/dotnetmentor/racoon/internal/output"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Json", func() {
	Describe("Write", func() {
		keys := []string{
			"Foo",
			"Bar",
			"CamelCasedProperty",
			"Path.Based.Property",
			"Dotnet.Structured.FormattedProperty",
		}
		values := map[string]string{
			"Foo":                                 "Bar",
			"Bar":                                 "Foo",
			"CamelCasedProperty":                  "Value",
			"Path.Based.Property":                 "Value",
			"Dotnet.Structured.FormattedProperty": "Value",
		}

		When("writing with defaults", func() {
			var result string

			BeforeEach(func() {
				_, stdout, _ := pio.Buffered(os.Stdin)
				o := output.NewJson()
				o.Write(stdout, keys, map[string]string{}, values)
				b, _ := io.ReadAll(stdout)
				result = string(b)
			})

			It("has structured output", func() {
				Expect(result).To(MatchJSON(`{"Bar":"Foo","CamelCasedProperty":"Value","Dotnet":{"Structured":{"FormattedProperty":"Value"}},"Foo":"Bar","Path":{"Based":{"Property":"Value"}}}`))
			})
		})

		When("writing unstructured json", func() {
			var result string

			BeforeEach(func() {
				_, stdout, _ := pio.Buffered(os.Stdin)
				o := output.NewJson()
				o.Stuctured = false
				o.Write(stdout, keys, map[string]string{}, values)
				b, _ := io.ReadAll(stdout)
				result = string(b)
			})

			It("has unstructured output", func() {
				Expect(result).To(MatchJSON(`{"Bar":"Foo","CamelCasedProperty":"Value","Dotnet.Structured.FormattedProperty":"Value","Foo":"Bar","Path.Based.Property":"Value"}`))
			})
		})
	})
})
