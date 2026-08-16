package catalog

import (
	"fmt"
	"mitre_red_team/internal/model"
)

// Validate 校验目录数据的一致性，返回首个错误。
// 覆盖：战术 ID 唯一、技术 ID 唯一、执行模式合法、技术所属战术存在。
// 战术声明的技术引用允许尚未实现（规划与实现分离，查询时只返回已实现项）。
func Validate(catalogData *Catalog) error {
	if err := validateTacticIDs(catalogData.Tactics); err != nil {
		return err
	}
	if err := validateTechniqueIDs(catalogData.Techniques); err != nil {
		return err
	}
	if err := validateTechniqueModes(catalogData.Techniques); err != nil {
		return err
	}
	return validateTechniqueReferences(catalogData)
}

// validateTacticIDs 检查战术 ID 是否重复。
func validateTacticIDs(tactics []model.Tactic) error {
	seen := make(map[string]struct{}, len(tactics))
	for _, tactic := range tactics {
		if _, exists := seen[tactic.ID]; exists {
			return fmt.Errorf("战术 ID 重复: %s", tactic.ID)
		}
		seen[tactic.ID] = struct{}{}
	}
	return nil
}

// validateTechniqueIDs 检查技术 ID 是否重复。
func validateTechniqueIDs(techniques []model.Technique) error {
	seen := make(map[string]struct{}, len(techniques))

	for _, technique := range techniques {
		if _, exists := seen[technique.ID]; exists {
			return fmt.Errorf("技术 ID 重复: %s", technique.ID)
		}
		seen[technique.ID] = struct{}{}
	}

	return nil
}

// validateTechniqueModes 检查执行模式是否属于 passive/active/manual。
func validateTechniqueModes(techniques []model.Technique) error {
	for _, technique := range techniques {
		switch technique.Mode {
		case model.TechniquePassive, model.TechniqueActive, model.TechniqueManual:
		default:
			return fmt.Errorf("非法的执行模式: %s（技术 %s）", technique.Mode, technique.ID)
		}
	}
	return nil
}

// validateTechniqueReferences 检查技术所属战术是否存在。
func validateTechniqueReferences(catalogData *Catalog) error {
	tacticIDs := make(map[string]struct{}, len(catalogData.Tactics))
	for _, tactic := range catalogData.Tactics {
		tacticIDs[tactic.ID] = struct{}{}
	}
	for _, technique := range catalogData.Techniques {
		if _, exists := tacticIDs[technique.TacticID]; !exists {
			return fmt.Errorf("技术 %s 引用了不存在的战术 %s", technique.ID, technique.TacticID)
		}
	}
	return nil
}
