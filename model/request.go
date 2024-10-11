package model

type Request struct {
	HasReturn bool `json:"has_return"`
}

var (
	NAME             = Request{HasReturn: true}
	ROLE             = Request{HasReturn: true}
	TALK             = Request{HasReturn: true}
	WHISPER          = Request{HasReturn: true}
	VOTE             = Request{HasReturn: true}
	DIVINE           = Request{HasReturn: true}
	GUARD            = Request{HasReturn: true}
	ATTACK           = Request{HasReturn: true}
	INITIALIZE       = Request{HasReturn: false}
	DAILY_INITIALIZE = Request{HasReturn: false}
	DAILY_FINISH     = Request{HasReturn: false}
	FINISH           = Request{HasReturn: false}
)
