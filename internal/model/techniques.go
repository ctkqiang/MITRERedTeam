package model

type Technique struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`

	Tactics   []string `json:"tactics" yaml:"tactics"`
	Platforms []string `json:"platforms" yaml:"platforms"`
	SubTechID string   `json:"subtechnique_id,omitempty" yaml:"subtechnique_id,omitempty"`

	Capabilities []string `json:"capabilities" yaml:"capabilities"`
	Tools        []string `json:"tools" yaml:"tools"`
}
