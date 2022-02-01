package output

import (
	"fmt"
	"strings"

	"github.com/dotnetmentor/racoon/internal/config"

	"github.com/fatih/camelcase"
)

func Dotenv(m config.Manifest, values map[string]string) string {
	var b strings.Builder
	for _, s := range m.Secrets {
		parts := camelcase.Split(s.Name)
		for i, part := range parts {
			parts[i] = strings.ToUpper(part)
		}
		key := strings.Join(parts, "_")
		value := strings.TrimSuffix(values[s.Name], "\n")
		fmt.Fprintf(&b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}
