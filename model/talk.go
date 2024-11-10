package model

import "encoding/json"

type Talk struct {
	Idx   int    `json:"idx"`
	Day   int    `json:"day"`
	Turn  int    `json:"turn"`
	Agent Agent  `json:"agent"`
	Text  string `json:"text"`
}

func (t Talk) MarshalJSON() ([]byte, error) {
	type Alias Talk
	return json.Marshal(&struct {
		*Alias
		Skip bool `json:"skip"`
		Over bool `json:"over"`
	}{
		Alias: (*Alias)(&t),
		Skip:  t.Text == T_SKIP || t.Text == T_FORCE_SKIP,
		Over:  t.Text == T_OVER,
	})
}

const (
	T_OVER       = "Over"
	T_SKIP       = "Skip"
	T_FORCE_SKIP = "ForceSkip"
)
