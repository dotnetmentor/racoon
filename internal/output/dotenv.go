package output

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/dotnetmentor/racoon/internal/utils"
)

type Dotenv struct {
	Sort          bool   `yaml:"sort"`
	Quote         bool   `yaml:"quote"`
	Uppercase     bool   `yaml:"uppercase"`
	WordSeparator string `yaml:"wordSeparator"`
	PathSeparator string `yaml:"pathSeparator"`
}

func NewDotenv() Dotenv {
	return Dotenv{
		Sort:          false,
		Quote:         true,
		Uppercase:     true,
		WordSeparator: "_",
		PathSeparator: "_",
	}
}

func (o Dotenv) Type() string {
	return "dotenv"
}

func (o Dotenv) Write(w io.Writer, keys []string, remap map[string]string, values map[string]string) {
	output := make(map[string]string)
	outputKeys := make([]string, len(keys))

	for i, k := range keys {
		var key string
		if remapped, ok := remap[k]; ok && remapped != "" {
			key = remapped
		} else {
			key = utils.FormatKey(k, utils.Formatting{
				Uppercase:     o.Uppercase,
				WordSeparator: o.WordSeparator,
				PathSeparator: o.PathSeparator,
			})
		}

		value := strings.TrimSuffix(values[k], "\n")
		format := "%s=%s\n"
		if o.Quote {
			format = "%s=\"%s\"\n"
		}
		output[key] = fmt.Sprintf(format, key, value)
		outputKeys[i] = key
	}

	if o.Sort {
		sort.Strings(outputKeys)
	}

	for _, k := range outputKeys {
		w.Write([]byte(output[k]))
	}
}
