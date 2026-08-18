package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"mitre_red_team/internal/agent"
	"mitre_red_team/internal/catalog"
	"mitre_red_team/internal/config"
	"mitre_red_team/internal/engine"
	"mitre_red_team/internal/llm"
	"mitre_red_team/internal/model"
	"mitre_red_team/internal/technique"
	"mitre_red_team/internal/technique/enumeration"
	"mitre_red_team/internal/utilities"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	techniqueID := flag.String("technique", "", "要执行的技术 ID，如 BB05.001")
	tacticID := flag.String("tactic", "", "要执行的战术 ID，如 BB05")
	mitreID := flag.String("mitre", "", "要执行的 MITRE ATT&CK 技术 ID，如 T1046")
	targetURL := flag.String("url", "", "目标 URL，必填")
	configPath := flag.String("config", "configs/redteam.json", "配置文件路径")
	manualWordlist := flag.String("wordlist", "", "自定义字典文件路径（每行一个条目，UTF-8 编码）")
	flag.StringVar(manualWordlist, "w", "", "自定义字典文件路径的短别名")
	aiMode := flag.Bool("ai", false, "启用 AI 辅助：随机选择已配置的 LLM 供应商，分析 TTP 输出并自动执行建议的下一步技术（最多 3 轮）")

	// 应用全程使用结构化日志器，日志行含时间戳、级别、操作名与描述，便于诊断执行流程。
	logger := utilities.New(os.Stderr)

	// 加载 .env：文件不存在或条目未配置时静默忽略，只有已配置的项才进入进程环境。
	if err := utilities.LoadDotenv(".env"); err != nil {
		fail(logger, "LoadDotenv", err)
	}
	logger.Info("LoadDotenv", nil, "环境文件加载完成")

	configuration, err := config.Load(*configPath)
	if err != nil {
		fail(logger, "LoadConfig", err)
	}
	logger.Info("LoadConfig", nil, fmt.Sprintf("已加载配置 %s，工具 %d 项，字典 %d 项", *configPath, len(configuration.Tools), len(configuration.Wordlists)))

	catalogData, err := catalog.Load("catalog")
	if err != nil {
		fail(logger, "LoadCatalog", err)
	}
	if err := catalog.Validate(catalogData); err != nil {
		fail(logger, "ValidateCatalog", err)
	}
	logger.Info("LoadCatalog", nil, fmt.Sprintf("目录校验通过：战术 %d 条，技术 %d 条", len(catalogData.Tactics), len(catalogData.Techniques)))

	// 帮助菜单需要展示 catalog 中的全部 TTP，因此 usage 依赖已加载的目录数据。
	flag.Usage = func() { printUsage(catalogData) }
	flag.Parse()
	logger.Info("ParseFlags", nil, fmt.Sprintf("technique=%q tactic=%q mitre=%q url=%q", *techniqueID, *tacticID, *mitreID, *targetURL))

	if err := validateFlags(*techniqueID, *tacticID, *mitreID, *targetURL); err != nil {
		logger.Error("ValidateFlags", nil, err.Error())
		fmt.Fprintln(os.Stderr, "错误:", err)
		printUsage(catalogData)
		os.Exit(1)
	}
	logger.Info("ValidateFlags", nil, "命令行参数校验通过")

	if missingTools := configuration.CheckToolsAvailable(); len(missingTools) > 0 {
		logger.Error("CheckTools", nil, fmt.Sprintf("缺少依赖工具 %v", missingTools))
		fmt.Fprintln(os.Stderr, "错误: 缺少以下必需工具，请安装后重试:")
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
	logger.Info("CheckTools", nil, "全部依赖工具可用")

	wordlistPath, err := resolveWordlist(logger, configuration, *manualWordlist)
	if err != nil {
		fail(logger, "ResolveWordlist", err)
	}

	registerTechniques(configuration, wordlistPath)

	target, err := parseTarget(*targetURL)
	if err != nil {
		fail(logger, "ParseTarget", err)
	}
	logger.Info("ParseTarget", nil, fmt.Sprintf("目标解析完成：主机 %s，协议 %s，端口 %d", target.Host, target.Scheme, target.Port))

	executionEngine := engine.New(catalogData)

	var results []model.ExecutionResult
	var mode string
	switch {
	case *aiMode:
		mode = "AI 辅助执行"
		logger.Info("Execute", nil, "进入 AI 辅助执行模式")
		results, err = runAI(
			context.Background(),
			executionEngine,
			catalogData,
			target,
			*mitreID,
			*techniqueID,
			*tacticID,
		)
	case *mitreID != "":
		mode = "按 MITRE ID 执行"
		logger.Info("Execute", nil, fmt.Sprintf("按 MITRE ID %s 执行", *mitreID))
		results, err = executionEngine.ExecuteByMitre(context.Background(), target, *mitreID)
	case *tacticID != "":
		mode = "按战术执行"
		logger.Info("Execute", nil, fmt.Sprintf("按战术 %s 执行", *tacticID))
		results, err = executionEngine.Execute(context.Background(), model.ExecutionRequest{
			Target:   target,
			TacticID: *tacticID,
		})
	default:
		mode = "按技术执行"
		logger.Info("Execute", nil, fmt.Sprintf("按技术 %s 执行", *techniqueID))
		results, err = executionEngine.Execute(context.Background(), model.ExecutionRequest{
			Target:      target,
			TechniqueID: *techniqueID,
		})
	}

	if err != nil {
		logger.Error("Execute", nil, fmt.Sprintf("%s失败：%v", mode, err))
		fmt.Fprintln(os.Stderr, "错误:", err)
		fmt.Fprintln(os.Stderr, "提示: 请检查目标是否可达、依赖工具是否已安装、字典文件是否有效后重试。")
		os.Exit(1)
	}

	successCount := 0
	for _, result := range results {
		if result.Status == model.StatusSucceeded {
			successCount++
		}
		fmt.Printf("%s %s: %s\n", result.Status, result.TechniqueID, result.Summary)
	}
	logger.Info("Summary", nil, fmt.Sprintf("%s完成：%d/%d 项成功", mode, successCount, len(results)))
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

// fallbackWordlistPath 是配置缺省时内置的默认字典路径。
const fallbackWordlistPath = "configs/wordlists/common.txt"

// registerTechniques 注册已实现的技术到注册表，使用解析后的字典路径。
func registerTechniques(configuration *config.Config, wordlist string) {
	technique.Register("directory-enumeration",
		enumeration.NewDirectoryEnumeration(configuration.Tools["ffuf"], wordlist))
}

// fail 记录错误日志、输出用户可读的错误消息后退出。
// 日志行保留完整错误上下文供诊断；终端消息面向普通用户，只呈现结果与应对方式。
func fail(logger *utilities.Logger, operation string, err error) {
	logger.Error(operation, nil, err.Error())
	fmt.Fprintln(os.Stderr, "错误:", err)
	fmt.Fprintln(os.Stderr, "提示: 请根据上述错误信息调整参数后重试，或使用 --help 查看完整用法。")
	os.Exit(1)
}

// resolveWordlist 决定本次执行使用的字典路径。
// 优先使用 --wordlist/-w 指定的自定义字典；未指定时在交互终端询问；
// 非交互环境跳过询问直接使用默认字典。自定义字典不可用时询问是否回退到默认。
func resolveWordlist(logger *utilities.Logger, configuration *config.Config, manualWordlist string) (string, error) {
	defaultWordlist := configuration.Wordlists["common"]
	if defaultWordlist == "" {
		defaultWordlist = fallbackWordlistPath
	}
	if manualWordlist != "" {
		logger.Info("ResolveWordlist", nil, fmt.Sprintf("使用命令行指定字典 %s", manualWordlist))
	}
	wordlistPath, err := enumeration.ResolveWordlistPath(manualWordlist, defaultWordlist, os.Stdin, os.Stderr)
	if err == nil {
		logger.Info("ResolveWordlist", nil, fmt.Sprintf("选定字典 %s", wordlistPath))
		return wordlistPath, nil
	}
	logger.Error("ResolveWordlist", nil, fmt.Sprintf("自定义字典不可用：%v", err))
	if !enumeration.ConfirmDefaultFallback(os.Stdin, os.Stderr) {
		logger.Warn("ResolveWordlist", nil, "用户拒绝回退，执行中止")
		return "", errors.New("已取消执行")
	}
	if err := enumeration.ValidateWordlist(defaultWordlist); err != nil {
		return "", fmt.Errorf("默认字典同样不可用: %w", err)
	}
	logger.Info("ResolveWordlist", nil, fmt.Sprintf("回退到默认字典 %s", defaultWordlist))
	return defaultWordlist, nil
}

// runAI 以 AI 辅助模式执行：从已配置的 LLM 供应商中随机选择一家，
// 执行用户请求的初始技术，把执行输出交给 LLM 分析，并自动推进建议的下一步技术。
func runAI(
	ctx context.Context,
	executionEngine *engine.Engine,
	catalogData *catalog.Catalog,
	target model.Target,
	mitreID string,
	techniqueID string,
	tacticID string,
) ([]model.ExecutionResult, error) {
	initial := func(ctx context.Context) ([]model.ExecutionResult, error) {
		if mitreID != "" {
			return executionEngine.ExecuteByMitre(ctx, target, mitreID)
		}
		return executionEngine.Execute(ctx, model.ExecutionRequest{
			Target:      target,
			TechniqueID: techniqueID,
			TacticID:    tacticID,
		})
	}
	return agent.Run(ctx, agent.RunParams{
		Engine:      executionEngine,
		CatalogData: catalogData,
		Target:      target,
		Initial:     initial,
		Providers:   llm.AvailableProviders(os.Getenv),
		Logger:      utilities.Default,
	})
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

// printUsage 打印命令用法、字典说明、示例与 catalog 中的全部 TTP 列表。
func printUsage(catalogData *catalog.Catalog) {
	fmt.Fprintln(os.Stderr, "用法: mitre_red_team --url <目标> [--technique <技术ID> | --tactic <战术ID> | --mitre <MITRE ID>]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "字典说明:")
	fmt.Fprintln(os.Stderr, "  默认字典：未指定任何自定义字典时，使用配置 configs/redteam.json 中")
	fmt.Fprintln(os.Stderr, "    wordlists.common 指向的字典文件（缺省为 configs/wordlists/common.txt）。")
	fmt.Fprintln(os.Stderr, "  自定义字典：通过 -w 或 --wordlist 指定字典文件路径。字典须为 UTF-8 编码，")
	fmt.Fprintln(os.Stderr, "    每行一个条目；空行与 # 开头的注释行将被忽略。")
	fmt.Fprintln(os.Stderr, "  交互式询问：未提供 --wordlist 时程序会提示输入字典路径，直接回车则使用默认字典。")
	fmt.Fprintln(os.Stderr, "  校验规则：自定义字典必须存在、可读且包含至少一条有效条目；")
	fmt.Fprintln(os.Stderr, "    校验失败时会提示原因，并可选择回退到默认字典继续执行。")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "示例:")
	fmt.Fprintln(os.Stderr, "  mitre_red_team --url https://example.com --technique BB05.001")
	fmt.Fprintln(os.Stderr, "  mitre_red_team --url https://example.com --tactic BB05")
	fmt.Fprintln(os.Stderr, "  mitre_red_team --url https://example.com --mitre T1046")
	fmt.Fprintln(os.Stderr, "  mitre_red_team --url https://example.com --technique BB05.001 -w /path/to/custom.txt")
	fmt.Fprintln(os.Stderr)
	printAvailableTTPs(catalogData)
	flag.PrintDefaults()
}

// printAvailableTTPs 以纯文本形式列出 catalog 中的全部战术与技术，
// 每项附上目录中的简要描述，描述为空时只展示编号与名称。
func printAvailableTTPs(catalogData *catalog.Catalog) {
	fmt.Fprintln(os.Stderr, "可用战术与技术（来自 catalog/）:")
	for _, tactic := range catalogData.Tactics {
		tacticLine := fmt.Sprintf("  %s  %s", tactic.ID, tactic.Name)
		if strings.TrimSpace(tactic.Description) != "" {
			tacticLine += " — " + tactic.Description
		}
		fmt.Fprintln(os.Stderr, tacticLine)
		for _, technique := range catalogData.TechniquesByTactic(tactic.ID) {
			techniqueLine := fmt.Sprintf("    %s  %s", technique.ID, technique.Name)
			if strings.TrimSpace(technique.Description) != "" {
				techniqueLine += " — " + technique.Description
			}
			fmt.Fprintln(os.Stderr, techniqueLine)
		}
	}
}
