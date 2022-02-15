package output

import "io"

type Output interface {
	Type() string
	Write(w io.Writer, secrets []string, remap map[string]string, values map[string]string)
}
