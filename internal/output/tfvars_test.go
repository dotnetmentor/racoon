package output_test

import (
	"io"
	"os"
	"strings"

	pio "github.com/dotnetmentor/racoon/internal/io"
	"github.com/dotnetmentor/racoon/internal/output"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tfvars", func() {
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
				o := output.NewTfvars()
				o.Write(stdout, keys, map[string]string{}, values)
				b, _ := io.ReadAll(stdout)
				result = string(b)
			})

			It("has qouted property", func() {
				Expect(result).To(ContainSubstring("foo = \"Bar\""))
			})

			It("keeps the sort order", func() {
				lines := strings.Split(result, "\n")
				Expect(lines[0]).To(ContainSubstring("foo ="))
				Expect(lines[1]).To(ContainSubstring("bar ="))
				Expect(lines[2]).To(ContainSubstring("camel_cased_property ="))
				Expect(lines[3]).To(ContainSubstring("path_based_property ="))
				Expect(lines[4]).To(ContainSubstring("dotnet_structured_formatted_property ="))
			})
		})

		When("writing without lowercase", func() {
			var result string

			BeforeEach(func() {
				_, stdout, _ := pio.Buffered(os.Stdin)
				o := output.NewTfvars()
				o.Lowercase = false
				o.Write(stdout, keys, map[string]string{}, values)
				b, _ := io.ReadAll(stdout)
				result = string(b)
			})

			It("keeps the casing", func() {
				lines := strings.Split(result, "\n")
				Expect(lines[0]).To(ContainSubstring("Foo ="))
				Expect(lines[1]).To(ContainSubstring("Bar ="))
				Expect(lines[2]).To(ContainSubstring("Camel_Cased_Property ="))
				Expect(lines[3]).To(ContainSubstring("Path_Based_Property ="))
				Expect(lines[4]).To(ContainSubstring("Dotnet_Structured_Formatted_Property ="))
			})
		})

		When("writing without a word separator", func() {
			var result string

			BeforeEach(func() {
				_, stdout, _ := pio.Buffered(os.Stdin)
				o := output.NewTfvars()
				o.Lowercase = false
				o.WordSeparator = ""
				o.Write(stdout, keys, map[string]string{}, values)
				b, _ := io.ReadAll(stdout)
				result = string(b)
			})

			It("keeps the casing", func() {
				lines := strings.Split(result, "\n")
				Expect(lines[0]).To(ContainSubstring("Foo ="))
				Expect(lines[1]).To(ContainSubstring("Bar ="))
				Expect(lines[2]).To(ContainSubstring("CamelCasedProperty ="))
				Expect(lines[3]).To(ContainSubstring("Path_Based_Property ="))
				Expect(lines[4]).To(ContainSubstring("Dotnet_Structured_FormattedProperty ="))
			})
		})
	})
})
