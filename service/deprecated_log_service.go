package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kano-lab/aiwolf-nlp-server/model"
)

type DeprecatedLogService struct {
	deprecatedLogsData map[string]*DeprecatedLogData
	outputDir          string
	templateFilename   string
}

type DeprecatedLogData struct {
	id       string
	filename string
	agents   []interface{}
	logs     []string
}

func NewDeprecatedLogService(config model.Config) *DeprecatedLogService {
	return &DeprecatedLogService{
		deprecatedLogsData: make(map[string]*DeprecatedLogData),
		outputDir:          config.DeprecatedLogService.OutputDir,
		templateFilename:   config.DeprecatedLogService.Filename,
	}
}

func (d *DeprecatedLogService) TrackStartGame(id string, agents []*model.Agent) {
	deprecatedLogData := &DeprecatedLogData{
		id:   id,
		logs: make([]string, 0),
	}
	for _, agent := range agents {
		deprecatedLogData.agents = append(deprecatedLogData.agents,
			map[string]interface{}{
				"idx":  agent.Idx,
				"team": agent.Team,
				"name": agent.Name,
				"role": agent.Role,
			},
		)
	}
	filename := strings.ReplaceAll(d.templateFilename, "{game_id}", deprecatedLogData.id)
	filename = strings.ReplaceAll(filename, "{timestamp}", fmt.Sprintf("%d", time.Now().Unix()))
	teams := make(map[string]struct{})
	for _, agent := range deprecatedLogData.agents {
		team := agent.(map[string]interface{})["team"].(string)
		teams[team] = struct{}{}
	}
	teamStr := ""
	for team := range teams {
		if teamStr != "" {
			teamStr += "_"
		}
		teamStr += team
	}
	filename = strings.ReplaceAll(filename, "{teams}", teamStr)
	deprecatedLogData.filename = filename

	d.deprecatedLogsData[id] = deprecatedLogData
}

func (d *DeprecatedLogService) TrackEndGame(id string) {
	if _, exists := d.deprecatedLogsData[id]; exists {
		d.saveDeprecatedLog(id)
		delete(d.deprecatedLogsData, id)
	}
}

func (d *DeprecatedLogService) AppendLog(id string, log string) {
	if deprecatedLogData, exists := d.deprecatedLogsData[id]; exists {
		deprecatedLogData.logs = append(deprecatedLogData.logs, log)
		d.saveDeprecatedLog(id)
	}
}

func (d *DeprecatedLogService) saveDeprecatedLog(id string) {
	if deprecatedLogData, exists := d.deprecatedLogsData[id]; exists {
		str := strings.Join(deprecatedLogData.logs, "\n")
		if _, err := os.Stat(d.outputDir); os.IsNotExist(err) {
			os.MkdirAll(d.outputDir, 0755)
		}
		filePath := filepath.Join(d.outputDir, fmt.Sprintf("%s.log", deprecatedLogData.filename))
		file, err := os.Create(filePath)
		if err != nil {
			return
		}
		defer file.Close()
		file.WriteString(str)
	}
}
