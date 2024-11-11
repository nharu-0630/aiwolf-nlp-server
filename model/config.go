package model

import (
	"log/slog"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		WebSocket struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"web_socket"`
		SelfMatch bool `yaml:"self_match"`
	} `yaml:"server"`
	Game struct {
		AgentCount            int     `yaml:"agent_count"`
		VoteVisibility        bool    `yaml:"vote_visibility"`
		TalkOnFirstDay        bool    `yaml:"talk_on_first_day"`
		MaxContinueErrorRatio float64 `yaml:"max_continue_error_ratio"`
		Talk                  struct {
			MaxCount struct {
				PerAgent int `yaml:"per_agent"`
				PerDay   int `yaml:"per_day"`
			} `yaml:"max_count"`
		} `yaml:"talk"`
		Whisper struct {
			MaxCount struct {
				PerAgent int `yaml:"per_agent"`
				PerDay   int `yaml:"per_day"`
			} `yaml:"max_count"`
		} `yaml:"whisper"`
		Skip struct {
			MaxCount int `yaml:"max_count"`
		} `yaml:"skip"`
		Vote struct {
			MaxCount int `yaml:"max_count"`
		} `yaml:"vote"`
		Attack struct {
			MaxCount      int  `yaml:"max_count"`
			AllowNoTarget bool `yaml:"allow_no_target"`
		} `yaml:"attack"`
		Timeout struct {
			Action   time.Duration `yaml:"action"`
			Response time.Duration `yaml:"response"`
		} `yaml:"timeout"`
	} `yaml:"game"`
	AnalysisService struct {
		OutputDir string `yaml:"output_dir"`
	} `yaml:"analysis_service"`
	ApiService struct {
		Enable             bool `yaml:"enable"`
		PublishRunningGame bool `yaml:"publish_running_game"`
	} `yaml:"api_service"`
	MatchOptimizer struct {
		Enable     bool   `yaml:"enable"`
		TeamCount  int    `yaml:"team_count"`
		GameCount  int    `yaml:"game_count"`
		OutputPath string `yaml:"output_path"`
	} `yaml:"match_optimizer"`
}

const WebSocketExternalHost = "0.0.0.0"

func LoadFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
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
