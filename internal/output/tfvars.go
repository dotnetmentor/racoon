package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/dotnetmentor/racoon/internal/config"

	"github.com/fatih/camelcase"
)

func Tfvars(w io.Writer, m config.Manifest, values map[string]string) {
	for _, s := range m.Secrets {
		parts := camelcase.Split(s.Name)
		for i, part := range parts {
			parts[i] = strings.ToLower(part)
		}
		key := strings.Join(parts, "_")
		value := strings.TrimSuffix(values[s.Name], "\n")
		w.Write([]byte(fmt.Sprintf("%s = \"%s\"\n", key, value)))
	}
}
