package model

import (
	"log/slog"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	WebSocket struct {
		Host string `yaml:"host"` // ホスト名
		Port int    `yaml:"port"` // ポート番号
	} `yaml:"web_socket"`

	Server struct {
		SelfMatch bool `yaml:"self_match"` // 同じチーム名のエージェント同士のみをマッチングさせるかどうか
	} `yaml:"server"`

	Game struct {
		AgentCount             int     `yaml:"agent_count"`                // 1ゲームあたりのエージェント数
		VoteVisibility         bool    `yaml:"vote_visibility"`            // 投票の結果を公開するか
		TalkOnFirstDay         bool    `yaml:"talk_on_first_day"`          // 1日目の発言を許可するか
		MaxHasErrorAgentsRatio float64 `yaml:"max_has_error_agents_ratio"` // ゲームを継続するエラーエージェントの最大割合
		Talk                   struct {
			MaxCount struct {
				PerAgent int `yaml:"per_agent"` // 1日あたりの1エージェントの最大発言回数
				PerDay   int `yaml:"per_day"`   // 1日あたりの全体の発言回数
			} `yaml:"max_count"`
		} `yaml:"talk"`
		Whisper struct {
			MaxCount struct {
				PerAgent int `yaml:"per_agent"` // 1日あたりの1エージェントの最大囁き回数
				PerDay   int `yaml:"per_day"`   // 1日あたりの全体の囁き回数
			} `yaml:"max_count"`
		} `yaml:"whisper"`
		Skip struct {
			MaxCount int `yaml:"max_count"` // 1日あたりの1エージェントの最大スキップ回数
		} `yaml:"skip"`
		Vote struct {
			MaxCount int `yaml:"max_count"` // 1位タイの場合の最大再投票回数
		} `yaml:"vote"`
		Attack struct {
			MaxCount      int  `yaml:"max_count"`       // 1位タイの場合の最大襲撃再投票回数
			AllowNoTarget bool `yaml:"allow_no_target"` // 襲撃なしの日を許可するか
		} `yaml:"attack"`
		Timeout struct {
			Action   time.Duration `yaml:"action"`   // エージェントのアクションのタイムアウト時間
			Response time.Duration `yaml:"response"` // エージェントの生存確認のタイムアウト時間
		} `yaml:"timeout"`
	} `yaml:"game"`

	AnalysisService struct {
		OutputDir string `yaml:"output_dir"` // 分析結果の出力ディレクトリ
	} `yaml:"analysis_service"`

	ApiService struct {
		Enable             bool `yaml:"enable"`               // APIサービスの有効化
		PublishRunningGame bool `yaml:"publish_running_game"` // 進行中のゲームを公開するか
	} `yaml:"api_service"`

	MatchOptimizer struct {
		OutputPath string `yaml:"output_path"` // マッチ履歴の出力先
	} `yaml:"match_optimizer"`
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
