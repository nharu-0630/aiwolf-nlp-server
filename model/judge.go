package model

type Judge struct {
	Day    int     `json:"day"`
	Agent  Agent   `json:"agent"`
	Target Agent   `json:"target"`
	Result Species `json:"result"`
}
