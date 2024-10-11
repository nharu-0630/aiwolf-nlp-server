package model

type Talk struct {
	Idx   int    `json:"idx"`
	Day   int    `json:"day"`
	Turn  int    `json:"turn"`
	Agent Agent  `json:"agent"`
	Text  string `json:"text"`
}

const (
	OVER       = "Over"
	SKIP       = "Skip"
	FORCE_SKIP = "ForceSkip"
)

func (t *Talk) IsOver() bool {
	return t.Text == OVER
}

func (t *Talk) IsSkip() bool {
	return t.Text == SKIP
}
