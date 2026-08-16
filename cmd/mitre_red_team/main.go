package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"mitre_red_team/internal/catalog"
	"mitre_red_team/internal/config"
	"mitre_red_team/internal/engine"
	"mitre_red_team/internal/model"
	"mitre_red_team/internal/technique"
	"mitre_red_team/internal/technique/enumeration"
)

func main() {
	techniqueID := flag.String("technique", "", "要执行的技术 ID，如 BB05.001")
	tacticID := flag.String("tactic", "", "要执行的战术 ID，如 BB05")
	mitreID := flag.String("mitre", "", "要执行的 MITRE ATT&CK 技术 ID，如 T1046")
	url := flag.String("url", "", "目标 URL，必填")
	configPath := flag.String("config", "configs/redteam.json", "配置文件路径")
	flag.Usage = printUsage
	flag.Parse()

	if err := validateFlags(*techniqueID, *tacticID, *mitreID, *url); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		printUsage()
		os.Exit(1)
	}

	configuration, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
	catalogData, err := catalog.Load("catalog")
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
	if err := catalog.Validate(catalogData); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}

	if missingTools := configuration.CheckToolsAvailable(); len(missingTools) > 0 {
		fmt.Fprintln(os.Stderr, "错误: 缺少以下必需工具:")
		for _, name := range missingTools {
			fmt.Fprintf(os.Stderr, "  - %s\n", name)
			if hint, known := toolInstallHints[name]; known {
				fmt.Fprintf(os.Stderr, "    安装: %s\n", hint)
			} else {
				fmt.Fprintln(os.Stderr, "    安装: 请查阅该工具的官方文档")
			}
		}
		os.Exit(1)
	}

	registerTechniques(configuration)

	target, err := parseTarget(*url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}

	var results []model.ExecutionResult
	executionEngine := engine.New(catalogData)
	if *mitreID != "" {
		results, err = executionEngine.ExecuteByMitre(context.Background(), target, *mitreID)
	} else {
		results, err = executionEngine.Execute(context.Background(), model.ExecutionRequest{
			Target:      target,
			TechniqueID: *techniqueID,
			TacticID:    *tacticID,
		})
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}

	for _, result := range results {
		fmt.Printf("%s %s: %s\n", result.Status, result.TechniqueID, result.Summary)
	}
}

// toolInstallHints 提供缺失工具的安装指引。
var toolInstallHints = map[string]string{
	"ffuf":      "brew install ffuf 或参考 https://github.com/ffuf/ffuf",
	"nmap":      "brew install nmap",
	"nuclei":    "go install github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest",
	"httpx":     "go install github.com/projectdiscovery/httpx/cmd/httpx@latest",
	"subfinder": "go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest",
	"sqlmap":    "brew install sqlmap",
}

// registerTechniques 注册已实现的技术到注册表。
func registerTechniques(configuration *config.Config) {
	const defaultWordlist = "common.txt"
	technique.Register("directory-enumeration",
		enumeration.NewDirectoryEnumeration(configuration.Tools["ffuf"], defaultWordlist))
}

// validateFlags 校验命令行参数：目标必填，技术/战术/MITRE 三者选一。
func validateFlags(techniqueID string, tacticID string, mitreID string, targetURL string) error {
	if targetURL == "" {
		return fmt.Errorf("必须提供 --url 目标")
	}
	selectors := 0
	for _, value := range []string{techniqueID, tacticID, mitreID} {
		if value != "" {
			selectors++
		}
	}
	switch selectors {
	case 0:
		return fmt.Errorf("必须提供 --technique、--tactic 或 --mitre")
	case 1:
		return nil
	default:
		return fmt.Errorf("--technique、--tactic、--mitre 只能选其一")
	}
}

// parseTarget 从 URL 解析目标的主机、协议与端口。
func parseTarget(rawURL string) (model.Target, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return model.Target{}, fmt.Errorf("解析目标 URL %s: %w", rawURL, err)
	}
	target := model.Target{
		Host:   parsed.Hostname(),
		Scheme: parsed.Scheme,
	}
	if port := parsed.Port(); port != "" {
		parsedPort, err := strconv.Atoi(port)
		if err != nil {
			return model.Target{}, fmt.Errorf("解析端口 %s: %w", port, err)
		}
		target.Port = parsedPort
	}
	return target, nil
}

// printUsage 打印命令用法与示例。
func printUsage() {
	fmt.Fprintln(os.Stderr, "用法: mitre_red_team --url <目标> [--technique <技术ID> | --tactic <战术ID> | --mitre <MITRE ID>]")
	fmt.Fprintln(os.Stderr, "示例:")
	fmt.Fprintln(os.Stderr, "  mitre_red_team --url https://example.com --technique BB05.001")
	fmt.Fprintln(os.Stderr, "  mitre_red_team --url https://example.com --tactic BB05")
	fmt.Fprintln(os.Stderr, "  mitre_red_team --url https://example.com --mitre T1046")
	flag.PrintDefaults()
}
