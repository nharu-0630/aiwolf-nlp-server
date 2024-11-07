package model

import (
	"log/slog"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	WebSocketHost string `yaml:"web_socket.host"` // ホスト名
	WebSocketPort int    `yaml:"web_socket.port"` // ポート番号

	AgentCount              int           `yaml:"game.agent_count"`                 // 1ゲームあたりのエージェント数
	AllowNoAttack           bool          `yaml:"game.allow_no_attack"`             // 襲撃なしの日を許可するか
	VoteVisibility          bool          `yaml:"game.vote_visibility"`             // 投票の結果を公開するか
	TalkOnFirstDay          bool          `yaml:"game.talk_on_first_day"`           // 1日目の発言を許可するか
	MaxHasErrorAgentsRatio  float64       `yaml:"game.max_has_error_agents_ratio"`  // ゲームを継続するエラーエージェントの最大割合
	MaxTalkCountPerAgent    int           `yaml:"game.talk.max_count.per_agent"`    // 1日あたりの1エージェントの最大発言回数
	MaxTalkCountPerDay      int           `yaml:"game.talk.max_count.per_day"`      // 1日あたりの全体の発言回数
	MaxWhisperCountPerAgent int           `yaml:"game.whisper.max_count.per_agent"` // 1日あたりの1エージェントの最大囁き回数
	MaxWhisperCountPerDay   int           `yaml:"game.whisper.max_count.per_day"`   // 1日あたりの全体の囁き回数
	MaxSkipCount            int           `yaml:"game.skip.max_count"`              // 1日あたりの1エージェントの最大スキップ回数
	MaxVoteCount            int           `yaml:"game.vote.max_count"`              // 1位タイの場合の最大再投票回数
	MaxAttackCount          int           `yaml:"game.attack.max_count"`            // 1位タイの場合の最大襲撃再投票回数
	ActionTimeout           time.Duration `yaml:"game.timeout.action"`              // エージェントのアクションのタイムアウト時間
	ResponseTimeout         time.Duration `yaml:"game.timeout.response"`            // エージェントの生存確認のタイムアウト時間

	AnalysisServiceOutputDir string `yaml:"analysis_service.output_dir"` // 分析結果の出力ディレクトリ
	MatchOptimizerOutputPath string `yaml:"match_optimizer.output_path"` // マッチ履歴の出力先
}

var DefaultConfig = Config{
	WebSocketHost:            "127.0.0.1",
	WebSocketPort:            8080,
	AgentCount:               5,
	MaxTalkCountPerAgent:     3,
	MaxTalkCountPerDay:       15,
	MaxWhisperCountPerAgent:  3,
	MaxWhisperCountPerDay:    15,
	MaxSkipCount:             3,
	AllowNoAttack:            true,
	VoteVisibility:           false,
	TalkOnFirstDay:           true,
	ActionTimeout:            time.Duration(60) * time.Second,
	ResponseTimeout:          time.Duration(90) * time.Second,
	MaxVoteCount:             1,
	MaxAttackCount:           1,
	MaxHasErrorAgentsRatio:   0.2,
	AnalysisServiceOutputDir: "./../log",
	MatchOptimizerOutputPath: "./../log/match_optimizer.json",
}

const WebSocketExternalHost = "0.0.0.0"

func LoadConfigFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		slog.Error("設定ファイルの読み込みに失敗しました", "error", err)
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		slog.Error("設定ファイルのパースに失敗しました", "error", err)
		return nil, err
	}
	return &config, nil
}
