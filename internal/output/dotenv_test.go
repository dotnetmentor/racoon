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

var _ = Describe("Dotenv", func() {
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
				o := output.NewDotenv()
				o.Write(stdout, keys, map[string]string{}, values)
				b, _ := io.ReadAll(stdout)
				result = string(b)
			})

			It("has qouted property", func() {
				Expect(result).To(ContainSubstring("FOO=\"Bar\""))
			})

			It("keeps the sort order", func() {
				lines := strings.Split(result, "\n")
				Expect(lines[0]).To(ContainSubstring("FOO="))
				Expect(lines[1]).To(ContainSubstring("BAR="))
				Expect(lines[2]).To(ContainSubstring("CAMEL_CASED_PROPERTY="))
				Expect(lines[3]).To(ContainSubstring("PATH_BASED_PROPERTY="))
				Expect(lines[4]).To(ContainSubstring("DOTNET_STRUCTURED_FORMATTED_PROPERTY="))
			})
		})

		When("writing without quoutes", func() {
			var result string

			BeforeEach(func() {
				_, stdout, _ := pio.Buffered(os.Stdin)
				o := output.NewDotenv()
				o.Quote = false
				o.Write(stdout, keys, map[string]string{}, values)
				b, _ := io.ReadAll(stdout)
				result = string(b)
			})

			It("has unquoted property", func() {
				Expect(result).To(ContainSubstring("FOO=Bar"))
			})
		})

		When("writing with sorting enabled", func() {
			var result string

			BeforeEach(func() {
				_, stdout, _ := pio.Buffered(os.Stdin)
				o := output.NewDotenv()
				o.Sort = true
				o.Write(stdout, keys, map[string]string{}, values)
				b, _ := io.ReadAll(stdout)
				result = string(b)
			})

			It("sorts the output", func() {
				lines := strings.Split(result, "\n")
				Expect(lines[0]).To(ContainSubstring("BAR="))
				Expect(lines[1]).To(ContainSubstring("CAMEL_CASED_PROPERTY="))
				Expect(lines[2]).To(ContainSubstring("DOTNET_STRUCTURED_FORMATTED_PROPERTY="))
				Expect(lines[3]).To(ContainSubstring("FOO="))
				Expect(lines[4]).To(ContainSubstring("PATH_BASED_PROPERTY="))
			})
		})

		When("writing with path separator", func() {
			var result string

			BeforeEach(func() {
				_, stdout, _ := pio.Buffered(os.Stdin)
				o := output.NewDotenv()
				o.PathSeparator = "__"
				o.Write(stdout, keys, map[string]string{}, values)
				b, _ := io.ReadAll(stdout)
				result = string(b)
			})

			It("sorts the output", func() {
				Expect(result).To(ContainSubstring("PATH__BASED__PROPERTY="))
				Expect(result).To(ContainSubstring("DOTNET__STRUCTURED__FORMATTED_PROPERTY="))
			})
		})

		When("writing dotnet style output", func() {
			var result string

			BeforeEach(func() {
				_, stdout, _ := pio.Buffered(os.Stdin)
				o := output.NewDotenv()
				o.Uppercase = false
				o.WordSeparator = ""
				o.PathSeparator = "__"
				o.Write(stdout, keys, map[string]string{}, values)
				b, _ := io.ReadAll(stdout)
				result = string(b)
			})

			It("sorts the output", func() {
				Expect(result).To(ContainSubstring("Path__Based__Property="))
				Expect(result).To(ContainSubstring("Dotnet__Structured__FormattedProperty="))
			})
		})

		When("writing prefixed output", func() {
			var result string

			BeforeEach(func() {
				_, stdout, _ := pio.Buffered(os.Stdin)
				o := output.NewDotenv()
				o.Prefix = "PREFIX_"
				o.Write(stdout, keys, map[string]string{}, values)
				b, _ := io.ReadAll(stdout)
				result = string(b)
			})

			It("sorts the output", func() {
				Expect(result).To(ContainSubstring("PREFIX_FOO="))
				Expect(result).To(ContainSubstring("PREFIX_BAR="))
				Expect(result).To(ContainSubstring("PREFIX_CAMEL_CASED_PROPERTY="))
				Expect(result).To(ContainSubstring("PREFIX_PATH_BASED_PROPERTY="))
				Expect(result).To(ContainSubstring("PREFIX_DOTNET_STRUCTURED_FORMATTED_PROPERTY="))
			})
		})
	})
})
