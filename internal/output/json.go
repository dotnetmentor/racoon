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
func (o Json) Write(w io.Writer, secrets []string, remap map[string]string, values map[string]string) {
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
