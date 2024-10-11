package model

import "encoding/json"

type Guard struct {
	Day    int   `json:"day"`    // 日付
	Agent  Agent `json:"agent"`  // 護衛したエージェント
	Target Agent `json:"target"` // 護衛対象のエージェント
}

func (g Guard) MarshalJSON() ([]byte, error) {
	if g == (Guard{}) {
		return json.Marshal(nil)
	}
	return json.Marshal(g)
}
