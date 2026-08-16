package tests

import (
	"path/filepath"
	"strings"
	"testing"

	"mitre_red_team/internal/catalog"
	"mitre_red_team/internal/model"
)

// 验证真实 catalog 数据的加载数量。
func TestLoadRealCatalog(t *testing.T) {
	catalogData, err := catalog.Load(filepath.Join("..", "catalog"))
	if err != nil {
		t.Fatalf("加载真实目录失败: %v", err)
	}
	if len(catalogData.Tactics) != 25 {
		t.Errorf("期望 25 个战术，实际 %d", len(catalogData.Tactics))
	}
	if len(catalogData.Techniques) != 27 {
		t.Errorf("期望 27 条技术，实际 %d", len(catalogData.Techniques))
	}
}

// 验证真实目录数据通过一致性校验。
func TestValidateRealCatalog(t *testing.T) {
	catalogData, err := catalog.Load(filepath.Join("..", "catalog"))
	if err != nil {
		t.Fatalf("加载真实目录失败: %v", err)
	}
	if err := catalog.Validate(catalogData); err != nil {
		t.Errorf("真实目录校验应通过，实际报错: %v", err)
	}
}

// 验证按 ID 查找战术，命中与未命中两种情况。
func TestGetTactic(t *testing.T) {
	catalogData, err := catalog.Load(filepath.Join("..", "catalog"))
	if err != nil {
		t.Fatalf("加载真实目录失败: %v", err)
	}
	tactic, found := catalogData.GetTactic("BB05")
	if !found {
		t.Fatal("期望命中 BB05")
	}
	if tactic.Name == "" {
		t.Error("命中后战术名称不应为空")
	}
	if _, found := catalogData.GetTactic("BB99"); found {
		t.Error("不存在的战术不应命中")
	}
}

// 验证按 ID 查找技术，命中与未命中两种情况。
func TestGetTechnique(t *testing.T) {
	catalogData, err := catalog.Load(filepath.Join("..", "catalog"))
	if err != nil {
		t.Fatalf("加载真实目录失败: %v", err)
	}
	technique, found := catalogData.GetTechnique("BB10.001")
	if !found {
		t.Fatal("期望命中 BB10.001")
	}
	if technique.Name == "" {
		t.Error("命中后技术名称不应为空")
	}
	if _, found := catalogData.GetTechnique("BB99.001"); found {
		t.Error("不存在的技术不应命中")
	}
}

// 验证按战术返回已实现技术，且保持声明顺序。
func TestTechniquesByTactic(t *testing.T) {
	catalogData, err := catalog.Load(filepath.Join("..", "catalog"))
	if err != nil {
		t.Fatalf("加载真实目录失败: %v", err)
	}
	techniques := catalogData.TechniquesByTactic("BB05")
	if len(techniques) != 2 {
		t.Fatalf("期望 BB05 返回 2 项已实现技术，实际 %d", len(techniques))
	}
	if techniques[0].ID != "BB05.001" || techniques[1].ID != "BB05.003" {
		t.Errorf("返回顺序与声明不一致: %s, %s", techniques[0].ID, techniques[1].ID)
	}
	if len(catalogData.TechniquesByTactic("BB99")) != 0 {
		t.Error("不存在的战术应返回空切片")
	}
}

// 验证按 MITRE ATT&CK ID 查询技术：一个 ID 可能映射到多条技术。
func TestTechniquesByMitreID(t *testing.T) {
	catalogData, err := catalog.Load(filepath.Join("..", "catalog"))
	if err != nil {
		t.Fatalf("加载真实目录失败: %v", err)
	}

	matches := catalogData.TechniquesByMitreID("T1046")
	if len(matches) < 2 {
		t.Fatalf("期望 T1046 至少映射 2 条技术，实际 %d", len(matches))
	}
	for _, technique := range matches {
		found := false
		for _, mapped := range technique.MITRE {
			if mapped == "T1046" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("技术 %s 未声明映射到 T1046", technique.ID)
		}
	}

	if len(catalogData.TechniquesByMitreID("T9999")) != 0 {
		t.Error("未匹配的 MITRE ID 应返回空切片")
	}
}

// 验证战术 ID 重复时校验报错。
func TestValidateDuplicateTacticID(t *testing.T) {
	catalogData := &catalog.Catalog{
		Tactics: []model.Tactic{
			{ID: "BB05", Name: "Web 应用枚举"},
			{ID: "BB05", Name: "重复战术"},
		},
	}
	assertValidateErrorContains(t, catalogData, "战术 ID 重复")
}

// 验证技术 ID 重复时校验报错。
func TestValidateDuplicateTechniqueID(t *testing.T) {
	catalogData := &catalog.Catalog{
		Techniques: []model.Technique{
			{ID: "BB05.001", TacticID: "BB05", Mode: model.TechniqueActive},
			{ID: "BB05.001", TacticID: "BB05", Mode: model.TechniqueActive},
		},
	}
	assertValidateErrorContains(t, catalogData, "技术 ID 重复")
}

// 验证非法执行模式时校验报错。
func TestValidateInvalidMode(t *testing.T) {
	catalogData := &catalog.Catalog{
		Techniques: []model.Technique{
			{ID: "BB05.001", TacticID: "BB05", Mode: "illegal"},
		},
	}
	assertValidateErrorContains(t, catalogData, "非法的执行模式")
}

// 验证技术引用了不存在的战术时校验报错。
func TestValidateOrphanTechnique(t *testing.T) {
	catalogData := &catalog.Catalog{
		Tactics: []model.Tactic{
			{ID: "BB05", Name: "Web 应用枚举"},
		},
		Techniques: []model.Technique{
			{ID: "BB05.001", TacticID: "BB99", Mode: model.TechniqueActive},
		},
	}
	assertValidateErrorContains(t, catalogData, "引用了不存在的战术")
}

// assertValidateErrorContains 校验 Validate 返回错误，且错误信息包含指定片段。
func assertValidateErrorContains(t *testing.T, catalogData *catalog.Catalog, expectedFragment string) {
	t.Helper()
	err := catalog.Validate(catalogData)
	if err == nil {
		t.Fatalf("期望校验报错（含 %q），实际无错误", expectedFragment)
	}
	if !strings.Contains(err.Error(), expectedFragment) {
		t.Errorf("错误信息应包含 %q，实际: %v", expectedFragment, err)
	}
}
