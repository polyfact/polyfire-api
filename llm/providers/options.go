package providers

type ProviderOptions struct {
	StopWords *[]string
}

type TokenUsage struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

type Result struct {
	Result     string     `json:"result"`
	TokenUsage TokenUsage `json:"token_usage"`
	Err        error
}
