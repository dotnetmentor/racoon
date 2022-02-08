package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/dotnetmentor/racoon/internal/config"
)

func Dotenv(w io.Writer, m config.Manifest, keys map[string]string, values map[string]string) {
	for _, s := range m.Secrets {
		var key string
		if remapped, ok := keys[s.Name]; ok && remapped != "" {
			key = remapped
		} else {
			key = CamelCaseSplitToUpperJoinByUnderscore(s.Name)
		}

		value := strings.TrimSuffix(values[s.Name], "\n")
		w.Write([]byte(fmt.Sprintf("%s=\"%s\"\n", key, value)))
	}
}
