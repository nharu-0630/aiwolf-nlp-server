package config

import "time"

const (
	WEBSOCKET_INTERNAL_HOST     = "127.0.0.1"                     // ネットワーク受信接続の警告を回避するため
	WEBSOCKET_EXTERNAL_HOST     = "0.0.0.0"                       // 外部接続を許可する場合
	WEBSOCKET_PORT              = 8080                            // ポート番号
	AGENT_COUNT_PER_GAME        = 5                               // 1ゲームあたりのエージェント数
	MAX_TALK_COUNT_PER_AGENT    = 3                               // 1日あたりの1エージェントの最大発言回数
	MAX_TALK_COUNT_PER_DAY      = 15                              // 1日あたりの全体の発言回数
	MAX_WHISPER_COUNT_PER_AGENT = 3                               // 1日あたりの1エージェントの最大囁き回数
	MAX_WHISPER_COUNT_PER_DAY   = 15                              // 1日あたりの全体の囁き回数
	MAX_SKIP_COUNT_PER_DAY      = 3                               // 1日あたりの最大スキップ回数
	ALLOW_NO_ATTACK             = true                            // 襲撃なしの日を許可するか
	VOTE_VISIBILITY             = false                           // 投票の結果を公開するか
	TALK_ON_FIRST_DAY           = true                            // 1日目の発言を許可するか
	ACTION_TIMEOUT              = time.Duration(60) * time.Second // エージェントのアクションのタイムアウト時間
	RESPONSE_TIMEOUT            = time.Duration(90) * time.Second // エージェントの生存確認のタイムアウト時間
	MAX_REVOTE_COUNT            = 1                               // 1位タイの場合の最大再投票回数
	MAX_ATTACK_REVOTE_COUNT     = 1                               // 1位タイの場合の最大襲撃再投票回数
	MAX_HAS_ERROR_AGENTS_RATIO  = 0.2                             // ゲームを継続するエラーエージェントの最大割合
	ANALYSIS_OUTPUT_DIR         = "./../log"                      // 分析結果の出力ディレクトリ
	MATCH_OPTIMIZER_PATH        = "./../log/match_optimizer.json" // マッチ履歴の出力先
)
