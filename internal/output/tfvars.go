package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/dotnetmentor/racoon/internal/utils"
)

type Tfvars struct {
	Lowercase     bool   `yaml:"lowercase"`
	WordSeparator string `yaml:"wordSeparator"`
	PathSeparator string `yaml:"pathSeparator"`
}

func (o Tfvars) Type() string {
	return "tfvars"
}

func NewTfvars() Tfvars {
	return Tfvars{
		Lowercase:     true,
		WordSeparator: "_",
		PathSeparator: "_",
	}
}

func (o Tfvars) Write(w io.Writer, keys []string, remap map[string]string, values map[string]string) {
	for _, k := range keys {
		var key string
		if remapped, ok := remap[k]; ok && remapped != "" {
			key = remapped
		} else {
			key = utils.FormatKey(k, utils.Formatting{
				Lowercase:     o.Lowercase,
				WordSeparator: o.WordSeparator,
				PathSeparator: o.PathSeparator,
			})
		}

		value := strings.TrimSuffix(values[k], "\n")
		w.Write([]byte(fmt.Sprintf("%s = \"%s\"\n", key, value)))
	}
}
