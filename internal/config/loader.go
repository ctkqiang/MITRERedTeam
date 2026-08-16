package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Load 从 path 加载运行配置。
// 使用 json.Decoder 流式解码，遵循内存约束；文件缺失或 JSON 非法时返回带路径上下文的错误。
func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开配置文件 %s: %w", path, err)
	}
	defer file.Close()

	var configuration Config
	if err := json.NewDecoder(file).Decode(&configuration); err != nil {
		return nil, fmt.Errorf("解析配置文件 %s: %w", path, err)
	}
	return &configuration, nil
}
