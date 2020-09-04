package filter

import (
	"time"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-ftb-client-go/ftb"
)

type Store struct {
	Storage map[string]*Model
}

func (s *Store) Create(newFilter *models.NewFilter) string {
	filter := &Model{
		Filter: &models.Filter{
			UniqueTimestamp: 0,
			LastUpdated:     time.Time{},
			Dataset:         newFilter.Dataset,
			InstanceID:      "",
			Dimensions:      newFilter.Dimensions,
			Downloads:       nil,
			Events:          nil,
			FilterID:        "12345",
			State:           "",
			Published:       nil,
			Links:           models.LinkMap{},
		},
		DisclosureControl: DisclosureControl{
			Status:         ftb.StatusOK,
			Dimension:      "",
			BlockedOptions: []string{},
		},
	}

	s.Storage[filter.FilterID] = filter
	return filter.FilterID
}

func (s *Store) Update(f *Model) {
	s.Storage[f.FilterID] = f
}

func (s *Store) GetByID(id string) *Model {
	v, _ := s.Storage[id]
	return v
}
