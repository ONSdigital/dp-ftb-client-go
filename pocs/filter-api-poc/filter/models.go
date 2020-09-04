package filter

import "github.com/ONSdigital/dp-filter-api/models"

type Model struct {
	*models.Filter
	DisclosureControl DisclosureControl `json:"disclosure_control"`
}

type DisclosureControl struct {
	Status         string   `json:"status"`
	Dimension      string   `json:"dimension"`
	BlockedOptions []string `json:"options"`
	BlockedCount   int      `json:"count"`
}
