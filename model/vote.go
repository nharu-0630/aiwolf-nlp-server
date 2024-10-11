package model

type Vote struct {
	Day    int   `json:"day"`    // 日付
	Agent  Agent `json:"agent"`  // 投票したエージェント
	Target Agent `json:"target"` // 投票対象のエージェント
}
