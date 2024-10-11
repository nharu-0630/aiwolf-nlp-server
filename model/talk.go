package model

type Talk struct {
	Idx   int    `json:"idx"`   // インデックス
	Day   int    `json:"day"`   // 日付
	Turn  int    `json:"turn"`  // ターン
	Agent Agent  `json:"agent"` // エージェント
	Text  string `json:"text"`  // 本文
}

const (
	T_OVER     = "Over"
	T_SKIP       = "Skip"
	T_FORCE_SKIP = "ForceSkip"
)

func (t *Talk) IsOver() bool {
	return t.Text == T_OVER
}

func (t *Talk) IsSkip() bool {
	return t.Text == T_SKIP
}
