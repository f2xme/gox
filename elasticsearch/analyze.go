package elasticsearch

// AnalyzeResponse 表示 analyze 响应。
type AnalyzeResponse struct {
	// Tokens 分词结果。
	Tokens []AnalyzeToken `json:"tokens"`
}

// AnalyzeToken 表示 analyze 返回的单个 token。
type AnalyzeToken struct {
	// Token 分词文本。
	Token string `json:"token"`
	// StartOffset 起始偏移。
	StartOffset int `json:"start_offset"`
	// EndOffset 结束偏移。
	EndOffset int `json:"end_offset"`
	// Type token 类型。
	Type string `json:"type"`
	// Position token 位置。
	Position int `json:"position"`
}
