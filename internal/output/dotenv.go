package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/dotnetmentor/racoon/internal/utils"
)

type Dotenv struct {
	Quote bool `yaml:"quote"`
}

func (o Dotenv) Type() string {
	return "dotenv"
}

func NewDotenv() Dotenv {
	return Dotenv{
		Quote: true,
	}
}

func (o Dotenv) Write(w io.Writer, keys []string, remap map[string]string, values map[string]string) {
	for _, k := range keys {
		var key string
		if remapped, ok := remap[k]; ok && remapped != "" {
			key = remapped
		} else {
			key = utils.CamelCaseSplitToUpperJoinByUnderscore(k)
		}

		value := strings.TrimSuffix(values[k], "\n")
		format := "%s=%s\n"
		if o.Quote {
			format = "%s=\"%s\"\n"
		}
		w.Write([]byte(fmt.Sprintf(format, key, value)))
	}
}
