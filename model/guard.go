package model

type Guard struct {
	Day    int   `json:"day"`
	Agent  Agent `json:"agent"`
	Target Agent `json:"target"`
}
