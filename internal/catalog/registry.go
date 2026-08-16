package catalog

import "mitre_red_team/internal/model"

// GetTactic 按 ID 查找战术，未找到时返回 false。
func (c *Catalog) GetTactic(id string) (model.Tactic, bool) {
	for _, tactic := range c.Tactics {
		if tactic.ID == id {
			return tactic, true
		}
	}
	return model.Tactic{}, false
}

// GetTechnique 按 ID 查找技术，未找到时返回 false。
func (c *Catalog) GetTechnique(id string) (model.Technique, bool) {
	for _, technique := range c.Techniques {
		if technique.ID == id {
			return technique, true
		}
	}
	return model.Technique{}, false
}

// TechniquesByTactic 返回指定战术下的技术列表，保持战术声明的顺序。
// 战术不存在或没有技术时返回空切片。
func (c *Catalog) TechniquesByTactic(tacticID string) []model.Technique {
	tactic, found := c.GetTactic(tacticID)
	if !found {
		return nil
	}
	techniques := make([]model.Technique, 0, len(tactic.Techniques))
	for _, techniqueID := range tactic.Techniques {
		technique, exists := c.GetTechnique(techniqueID)
		if exists {
			techniques = append(techniques, technique)
		}
	}
	return techniques
}

// TechniquesByMitreID 返回映射到指定 MITRE ATT&CK ID 的所有技术。
// 一个 MITRE ID 可能对应多条技术，保持目录中声明的顺序；未匹配时返回空切片。
func (c *Catalog) TechniquesByMitreID(mitreID string) []model.Technique {
	var matches []model.Technique
	for _, technique := range c.Techniques {
		for _, mapped := range technique.MITRE {
			if mapped == mitreID {
				matches = append(matches, technique)
				break
			}
		}
	}
	return matches
}
