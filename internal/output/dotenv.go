package output

import (
	"fmt"
	"io"
	"strings"
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

func (o Dotenv) Write(w io.Writer, secrets []string, remap map[string]string, values map[string]string) {
	for _, s := range secrets {
		var key string
		if remapped, ok := remap[s]; ok && remapped != "" {
			key = remapped
		} else {
			key = CamelCaseSplitToUpperJoinByUnderscore(s)
		}

		value := strings.TrimSuffix(values[s], "\n")
		format := "%s=%s\n"
		if o.Quote {
			format = "%s=\"%s\"\n"
		}
		w.Write([]byte(fmt.Sprintf(format, key, value)))
	}
}
