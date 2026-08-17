package catalog

import "mitre_red_team/internal/model"

// Catalog 聚合战术与技术数据，是引擎查询目录的入口。
// 加载阶段会为 ID 建立索引，避免查询时反复线性扫描。
type Catalog struct {
	Tactics    []model.Tactic
	Techniques []model.Technique

	tacticIndex    map[string]model.Tactic
	techniqueIndex map[string]model.Technique
}

// buildIndex 加载阶段统一建立 ID 索引，使后续按 ID 查询为 O(1)。
func (c *Catalog) buildIndex() {
	c.tacticIndex = make(map[string]model.Tactic, len(c.Tactics))
	for _, tactic := range c.Tactics {
		c.tacticIndex[tactic.ID] = tactic
	}
	c.techniqueIndex = make(map[string]model.Technique, len(c.Techniques))
	for _, technique := range c.Techniques {
		c.techniqueIndex[technique.ID] = technique
	}
}
