package codebook

type Datasets struct {
	Items []*Dataset `json:"items,omitempty"`
}

type Dataset struct {
	Name             string `json:"name,omitempty"`
	Description      string `json:"description,omitempty"`
	Size             int    `json:"size,omitempty"`
	RuleRootVariable string `json:"rule_root_variable,omitempty"`
	Digest           string `json:"digest,omitempty"`
}

type Codebook struct {
	Dataset  Dataset     `json:"dataset"`
	CodeBook []Dimension `json:"codebook"`
}

type Dimension struct {
	Name         string   `json:"name"`
	Codes        []string `json:"codes"`
	Label        string   `json:"label"`
	Labels       []string `json:"labels"`
	MapFrom      []string `json:"mapFrom"`
	MapFromCodes []string `json:"mapFromCodes"`
}

func (d *Dimension) GetLabelByCode(code string) string {
	for i, val := range d.Codes {
		if val == code {
			return d.Labels[i]
		}
	}

	return code
}

func (d *Dimension) GetCodeByLabel(label string) string {
	for i, val := range d.Labels {
		if val == label {
			return d.Codes[i]
		}
	}

	return label
}

