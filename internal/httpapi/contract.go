package httpapi

import "encoding/json"

// CompareCommand defines an HTTP response object
type CompareCommand struct {
	Command string `json:"command"`
	Left    string `json:"left"`
	Right   string `json:"right"`
}

// CompareCommandResponse defines an HTTP response object
type CompareCommandResponse struct {
	Left  *ExecutionResult `json:"left"`
	Right *ExecutionResult `json:"right"`
}

type ExecutionResult struct {
	Logs   string `json:"logs"`
	Result string `json:"result"`
}

type ConfigQueryResponse struct {
	Error   string            `json:"error,omitempty"`
	Items   []ConfigQueryItem `json:"items"`
	Total   int               `json:"total"`
	More    bool              `json:"more"`
	Filters []string          `json:"filters"`
}

type ConfigQueryItem struct {
	Matches   int             `json:"matches"`
	Path      string          `json:"path"`
	Encrypted bool            `json:"encrypted"`
	Data      json.RawMessage `json:"data"`
}

type ConfigDecryptCommand struct {
	Path string `json:"path"`
}

type ConfigDecryptCommandResponse struct {
	Error string          `json:"error,omitempty"`
	Data  json.RawMessage `json:"data"`
}
