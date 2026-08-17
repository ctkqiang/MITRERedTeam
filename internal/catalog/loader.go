package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"mitre_red_team/internal/model"
)

// Load 从 directory 目录加载 tactics.json 与 techniques.json 并聚合返回。
// 使用 json.Decoder 流式解码，避免整文件读入常驻内存。
func Load(directory string) (*Catalog, error) {
	tactics, err := loadTactics(filepath.Join(directory, "tactics.json"))
	if err != nil {
		return nil, err
	}
	techniques, err := loadTechniques(filepath.Join(directory, "techniques.json"))
	if err != nil {
		return nil, err
	}
	catalogData := &Catalog{Tactics: tactics, Techniques: techniques}
	catalogData.buildIndex()
	return catalogData, nil
}

// loadTactics 流式解码 tactics.json 为战术列表。
func loadTactics(path string) ([]model.Tactic, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开战术目录 %s: %w", path, err)
	}
	defer file.Close()

	var tactics []model.Tactic
	if err := json.NewDecoder(file).Decode(&tactics); err != nil {
		return nil, fmt.Errorf("解析战术目录 %s: %w", path, err)
	}
	return tactics, nil
}

// loadTechniques 流式解码 techniques.json 为技术列表。
func loadTechniques(path string) ([]model.Technique, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开技术目录 %s: %w", path, err)
	}
	defer file.Close()

	var techniques []model.Technique
	if err := json.NewDecoder(file).Decode(&techniques); err != nil {
		return nil, fmt.Errorf("解析技术目录 %s: %w", path, err)
	}
	return techniques, nil
}
