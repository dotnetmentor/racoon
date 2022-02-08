package output

import (
	"encoding/json"
	"io"
	"strings"
)

func Json(w io.Writer, secrets []string, remap map[string]string, values map[string]string) {
	jo := map[string]string{}
	for _, s := range secrets {
		var key string
		if remapped, ok := remap[s]; ok && remapped != "" {
			key = remapped
		} else {
			key = s
		}

		value := strings.TrimSuffix(values[s], "\n")
		jo[key] = value
	}
	json.NewEncoder(w).Encode(jo)
}
