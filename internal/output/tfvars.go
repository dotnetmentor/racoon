package output

import (
	"fmt"
	"io"
	"strings"
)

type Tfvars struct {
}

func (o Tfvars) Type() string {
	return "tfvars"
}

func NewTfvars() Tfvars {
	return Tfvars{}
}

func (o Tfvars) Write(w io.Writer, secrets []string, remap map[string]string, values map[string]string) {
	for _, s := range secrets {
		var key string
		if remapped, ok := remap[s]; ok && remapped != "" {
			key = remapped
		} else {
			key = CamelCaseSplitToLowerJoinByUnderscore(s)
		}

		value := strings.TrimSuffix(values[s], "\n")
		w.Write([]byte(fmt.Sprintf("%s = \"%s\"\n", key, value)))
	}
}
