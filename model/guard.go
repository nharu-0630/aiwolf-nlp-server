package model

type Guard struct {
	Day    int   `json:"day"`    // 日付
	Agent  Agent `json:"agent"`  // 護衛したエージェント
	Target Agent `json:"target"` // 護衛対象のエージェント
}
