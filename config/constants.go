package config

const (
	WEBSOCKET_HOST = "127.0.0.1" // ネットワーク受信接続の警告を回避するため
	// WEBSOCKET_HOSTNAME = "0.0.0.0" // 外部接続を許可する場合
	WEBSOCKET_PORT   = 8080
	GAME_AGENT_COUNT = 5

	MAX_TALK_COUNT_PER_AGENT    = 0
	MAX_TALK_COUNT              = 0
	MAX_WHISPER_COUNT_PER_AGENT = 0
	MAX_WHISPER_COUNT           = 0
	MAX_SKIP_COUNT              = 3
	IS_ENABLE_NO_ATTACK         = true
	IS_VOTE_VISIBLE             = false
	IS_TALK_ON_FIRST_DAY        = true
	RESPONSE_TIMEOUT            = 6000
	ACTION_TIMEOUT              = 3000
	MAX_REVOTE_COUNT            = 1
	MAX_ATTACK_REVOTE_COUNT     = 1
)
