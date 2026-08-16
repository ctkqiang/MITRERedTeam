package catalog

import "mitre_red_team/internal/model"

// Catalog 聚合战术与技术数据，是引擎查询目录的入口。
type Catalog struct {
	Tactics    []model.Tactic
	Techniques []model.Technique
}
