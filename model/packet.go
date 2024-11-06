package model

type Packet struct {
	Request        *Request  `json:"request"`                  // リクエスト
	Info           *Info     `json:"info,omitempty"`           // 情報
	Settings       *Settings `json:"setting,omitempty"`        // 設定
	TalkHistory    *[]Talk   `json:"talkHistory,omitempty"`    // 会話履歴
	WhisperHistory *[]Talk   `json:"whisperHistory,omitempty"` // 囁き履歴
}
