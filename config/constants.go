package config

import "time"

const (
	WEBSOCKET_HOST = "127.0.0.1" // ネットワーク受信接続の警告を回避するため
	// WEBSOCKET_HOST              = "0.0.0.0"                       // 外部接続を許可する場合
	WEBSOCKET_PORT              = 8080                            // ポート番号
	AGENT_COUNT_PER_GAME        = 5                               // ゲームあたりのエージェント数
	MAX_TALK_COUNT_PER_AGENT    = 0                               // 1日あたりの1エージェントの最大発言回数
	MAX_TALK_COUNT_PER_DAY      = 0                               // 1日あたりの全体の発言回数
	MAX_WHISPER_COUNT_PER_AGENT = 0                               // 1日あたりの1エージェントの最大囁き回数
	MAX_WHISPER_COUNT_PER_DAY   = 0                               // 1日あたりの全体の囁き回数
	MAX_SKIP_COUNT_PER_DAY      = 3                               // 1日あたりの最大スキップ回数
	ALLOW_NO_ATTACK             = true                            // 襲撃なしの日を許可するか
	VOTE_VISIBILITY             = false                           // 投票の結果を公開するか
	TALK_ON_FIRST_DAY           = true                            // 1日目の発言を許可するか
	RESPONSE_TIMEOUT            = time.Duration(90) * time.Second // タイムアウト時間
	ACTION_TIMEOUT              = time.Duration(60) * time.Second // タイムアウト時間
	MAX_REVOTE_COUNT            = 1                               // 最大再投票回数
	MAX_ATTACK_REVOTE_COUNT     = 1                               // 最大襲撃再投票回数
)