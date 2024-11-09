package model

type Talk struct {
	Idx   int    `json:"idx"`
	Day   int    `json:"day"`
	Turn  int    `json:"turn"`
	Agent Agent  `json:"agent"`
	Text  string `json:"text"`
}

const (
	T_OVER       = "Over"
	T_SKIP       = "Skip"
	T_FORCE_SKIP = "ForceSkip"
)
