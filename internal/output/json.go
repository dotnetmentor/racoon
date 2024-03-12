package output

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/dotnetmentor/racoon/internal/utils"
)

type Json struct {
	Stuctured bool `yaml:"structured"`
}

func NewJson() Json {
	return Json{
		Stuctured: true,
	}
}

func (o Json) Type() string {
	return "json"
}

func (o Json) Write(w io.Writer, keys []string, remap map[string]string, values map[string]string) {
	jo := make(Dict)
	for _, k := range keys {
		var keyParts []string
		if remapped, ok := remap[k]; ok && remapped != "" {
			keyParts = []string{remapped}
		} else {
			if o.Stuctured {
				keyParts = utils.SplitPath(k)
			} else {
				keyParts = []string{k}
			}
		}

		value := strings.TrimSuffix(values[k], "\n")
		set(jo, keyParts, value)
	}
	if err := json.NewEncoder(w).Encode(jo); err != nil {
		panic(err)
	}
}

type Dict map[string]interface{}

func set(d Dict, keys []string, value interface{}) {
	if len(keys) == 1 {
		d[keys[0]] = value
		return
	}
	v, ok := d[keys[0]]
	if !ok {
		v = Dict{}
		d[keys[0]] = v
	}
	set(v.(Dict), keys[1:], value)
}
