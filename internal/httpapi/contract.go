package httpapi

// CompareRequest defines an HTTP response object
type CompareRequest struct {
	Command string `json:"command"`
	Left    string `json:"left"`
	Right   string `json:"right"`
}

// CompareResponse defines an HTTP response object
type CompareResponse struct {
	Left  *ExecutionResult `json:"left"`
	Right *ExecutionResult `json:"right"`
}

type ExecutionResult struct {
	Logs   string `json:"logs"`
	Result string `json:"result"`
}
