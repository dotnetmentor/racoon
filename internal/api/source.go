package api

import "fmt"

const (
	SourceTypeAwsParameterStore SourceType = "awsParameterStore"
	SourceTypeDefault           SourceType = "default"
	SourceTypeEnvironment       SourceType = "env"
	SourceTypeFormatter         SourceType = "formatter"
	SourceTypeLiteral           SourceType = "literal"
	SourceTypeParameter         SourceType = "parameter"
)

type SourceType string

func (st SourceType) Writable() bool {
	switch st {
	case SourceTypeAwsParameterStore:
		return true
	default:
		return false
	}
}

func NewValueSource(l Layer, t SourceType) ValueSource {
	return ValueSource{
		layer:      l,
		sourceType: t,
	}
}

type ValueSource struct {
	layer      Layer
	sourceType SourceType
}

func (s ValueSource) String() string {
	return fmt.Sprintf("%s/%s", s.layer.Name, s.sourceType)
}

func (s ValueSource) Layer() Layer {
	return s.layer
}

func (s ValueSource) Type() SourceType {
	return s.sourceType
}

func (s ValueSource) Writable() bool {
	switch s.sourceType {
	case SourceTypeAwsParameterStore:
		return true
	default:
		return false
	}
}
