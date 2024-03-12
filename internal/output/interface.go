package output

import "io"

type Output interface {
	Type() string
	Write(w io.Writer, keys []string, remap map[string]string, values map[string]string)
}
