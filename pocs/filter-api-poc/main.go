package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ONSdigital/dp-filter-api/models"
	"github.com/ONSdigital/dp-ftb-client-go/ftb"
	"github.com/ONSdigital/dp-ftb-client-go/pocs/filter-api-poc/filter"
	dpHTTP "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

var (
	port = ":22100"

	ftbHost = fmt.Sprintf("http://%s:10100", os.Getenv("EC2_IP"))
	ftbCli  = ftb.NewClient(ftbHost, os.Getenv("AUTH_PROXY_TOKEN"), dpHTTP.DefaultClient)
)

type FilterStore interface {
	Create(m *models.NewFilter) string
	Update(f *filter.Model)
	GetByID(id string) *filter.Model
}

func main() {
	if err := run(); err != nil {
		log.Event(nil, "application error", log.ERROR, log.Error(err))
		os.Exit(1)
	}
}

func run() error {
	store := &filter.Store{
		Storage: make(map[string]*filter.Model, 0),
	}

	err := loadFilter(store)
	if err != nil {
		return err
	}

	r := mux.NewRouter()
	r.HandleFunc("/filters", postNewFilter(store)).Methods(http.MethodPost)
	r.HandleFunc("/filters/{filter_id}", getFilter(store)).Methods("GET")
	r.HandleFunc("/filters/{filter_id}/dimensions/{name}/options/{option}", addDimensionOption(store)).Methods(http.MethodPost)

	ctx := context.Background()
	log.Event(ctx, "start mock observation API", log.INFO, log.Data{"PORT": port})

	return http.ListenAndServe(port, r)
}

func loadFilter(store FilterStore) error {
	b, err := ioutil.ReadFile("/Users/dave/Development/go/ons/dp-ftb-client-go/pocs/filter-api-poc/json/newFilter.json")
	if err != nil {
		return err
	}

	var f models.NewFilter
	err = json.Unmarshal(b, &f)
	if err != nil {
		return err
	}

	store.Create(&f)
	return nil
}

func postNewFilter(store FilterStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Event(ctx, "handling new filter request", log.INFO)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var f models.NewFilter
		err = json.Unmarshal(body, &f)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		id := store.Create(&f)
		newFilter := store.GetByID(id)

		b, err := json.Marshal(newFilter)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Add("content-type", "application/json")
		w.Write(b)
	}
}

func getFilter(store FilterStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		filterID := vars["filter_id"]

		f := store.GetByID(filterID)
		if f == nil {
			http.Error(w, "filter not found", http.StatusNotFound)
			return
		}

		b, err := json.Marshal(f)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(b)
	}
}

func addDimensionOption(store FilterStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		filterID := vars["filter_id"]
		dimensionName := vars["name"]
		option := vars["option"]

		f := store.GetByID(filterID)
		if f == nil {
			http.Error(w, "filter not found", http.StatusNotFound)
			return
		}

		var dims []models.Dimension
		for _, d := range f.Dimensions {
			if strings.ToLower(d.Name) == strings.ToLower(dimensionName) {
				exists := false
				for _, i := range d.Options {
					if i == option {
						exists = true
						break
					}
				}

				if !exists {
					d.Options = append(d.Options, option)
				}

			}
			dims = append(dims, d)
		}

		f.Dimensions = dims

		err := updateDisclosureControlStatus(r.Context(), f)
		if err != nil {
			http.Error(w, "ftb query error", http.StatusInternalServerError)
			return
		}

		store.Update(f)

		b, err := json.Marshal(models.PublicDimensionOption{Links: nil, Option: option})
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(b)
	}
}

func updateDisclosureControlStatus(ctx context.Context, f *filter.Model) error {
	options := make([]ftb.DimensionOptions, 0)
	for _, d := range f.Dimensions {
		if len(d.Options) > 0 {
			options = append(options, ftb.DimensionOptions{Name: d.Name, Options: d.Options})
		}
	}

	query := ftb.Query{
		DatasetName:       f.Dataset.ID,
		DimensionsOptions: options,
		RootDimension:     "OA",
		Limit:             0,
	}

	result, err := ftbCli.Query(ctx, query)
	if err != nil {
		return err
	}

	f.DisclosureControl.Status = result.DisclosureControlDetails.Status
	f.DisclosureControl.Dimension = result.DisclosureControlDetails.Dimension
	f.DisclosureControl.BlockedCount = result.DisclosureControlDetails.BlockedCount

	return nil
}
