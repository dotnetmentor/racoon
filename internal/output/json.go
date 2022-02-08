package output

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/dotnetmentor/racoon/internal/config"
)

func Json(w io.Writer, m config.Manifest, keys map[string]string, values map[string]string) {
	jo := map[string]string{}
	for _, s := range m.Secrets {
		var key string
		if remapped, ok := keys[s.Name]; ok && remapped != "" {
			key = remapped
		} else {
			key = s.Name
		}

		value := strings.TrimSuffix(values[s.Name], "\n")
		jo[key] = value
	}
	json.NewEncoder(w).Encode(jo)
}
