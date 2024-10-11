package model

import "encoding/json"

type Request struct {
	Type             string `json:"request"`           // リクエスト
	RequiredResponse bool   `json:"required_response"` // 応答が必要か
}

var (
	R_NAME = Request{
		Type:             "NAME",
		RequiredResponse: true,
	}
	R_ROLE = Request{
		Type:             "ROLE",
		RequiredResponse: true,
	}
	R_TALK = Request{
		Type:             "TALK",
		RequiredResponse: true}
	R_WHISPER = Request{
		Type:             "WHISPER",
		RequiredResponse: true}
	R_VOTE = Request{
		Type:             "VOTE",
		RequiredResponse: true}
	R_DIVINE = Request{
		Type:             "DIVINE",
		RequiredResponse: true}
	R_GUARD = Request{
		Type:             "GUARD",
		RequiredResponse: true}
	R_ATTACK = Request{
		Type:             "ATTACK",
		RequiredResponse: true}
	R_INITIALIZE = Request{
		Type:             "INITIALIZE",
		RequiredResponse: false}
	R_DAILY_INITIALIZE = Request{
		Type:             "DAILY_INITIALIZE",
		RequiredResponse: false}
	R_DAILY_FINISH = Request{
		Type:             "DAILY_FINISH",
		RequiredResponse: false}
	R_FINISH = Request{
		Type:             "FINISH",
		RequiredResponse: false}
)

func (r Request) String() string {
	return r.Type
}

func (r Request) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r Request) IsEmpty() bool {
	return r.Type == ""
}
