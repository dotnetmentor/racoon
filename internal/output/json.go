package output

import (
	"encoding/json"
	"io"
	"strings"
)

type Json struct {
}

func (o Json) Type() string {
	return "json"
}

func NewJson() Json {
	return Json{}
}
func (o Json) Write(w io.Writer, keys []string, remap map[string]string, values map[string]string) {
	jo := map[string]string{}
	for _, k := range keys {
		var key string
		if remapped, ok := remap[k]; ok && remapped != "" {
			key = remapped
		} else {
			key = k
		}

		value := strings.TrimSuffix(values[k], "\n")
		jo[key] = value
	}
	json.NewEncoder(w).Encode(jo)
}
