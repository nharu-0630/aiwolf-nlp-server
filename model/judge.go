package model

type Judge struct {
	Day    int     `json:"day"`    // 日付
	Agent  Agent   `json:"agent"`  // 判定したエージェント
	Target Agent   `json:"target"` // 判定対象のエージェント
	Result Species `json:"result"` // 判定結果
}
