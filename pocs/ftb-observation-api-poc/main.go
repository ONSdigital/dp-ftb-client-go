package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-ftb-client-go/ftb"
	dpHTTP "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-observation-api/models"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

var (
	port = ":24500"

	ftbHost = fmt.Sprintf("http://%s:10100", os.Getenv("EC2_IP"))
	ftbCli  = ftb.NewClient(ftbHost, os.Getenv("AUTH_PROXY_TOKEN"), dpHTTP.DefaultClient)
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations", GetObservations).Methods(http.MethodGet)

	ctx := context.Background()
	log.Event(ctx, "start mock observation API", log.INFO, log.Data{"PORT": port})

	err := http.ListenAndServe(port, r)
	if err != nil {
		log.Event(nil, "application error", log.ERROR, log.Error(err))
	}
}

func GetObservations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Event(ctx, "get observations request", log.INFO, log.Data{
		"request": r.URL.String(),
	})

	vars := mux.Vars(r)
	datasetID := vars["dataset_id"]

	doc, err := getObservations(ctx, datasetID, r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	b, err := json.Marshal(doc)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(b)
}

func getObservations(ctx context.Context, datasetName string, queryParams url.Values) (*models.ObservationsDoc, error) {
	query := ftb.NewQuery(datasetName, "COUNTRY", queryParams)
	result, err := ftbCli.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// wildcard dimensions only
	dimensions := make(map[string]*models.DimensionObject, 0)
	for _, d := range query.DimensionsOptions {
		if len(d.Options) == 0 {
			dimensions[d.Name] = &models.DimensionObject{
				HRef:  "",
				ID:    d.Name,
				Label: d.Name,
			}
		}
	}

	observations := make([]models.Observation, 0)

	for _, row := range result.ObservationsTable.Rows {
		dims := make(map[string]*models.DimensionObject, 0)

		for i := 0; i < len(row)-1; i++ {
			dims[result.ObservationsTable.Header[i]] = &models.DimensionObject{
				HRef:  "",
				ID:    row[i],
				Label: row[i],
			}
		}

		ob := models.Observation{
			Dimensions:  dims,
			Metadata:    nil,
			Observation: row[len(row)-1],
		}
		observations = append(observations, ob)
	}

	dimensionsMap := make(map[string]models.Option, 0)
	for _, d := range query.DimensionsOptions {
		if len(d.Options) > 0 {
			dimensionsMap[d.Name] = models.Option{
				LinkObject: &dataset.Link{
					URL: fmt.Sprintf("%s/v6/codebook/%s?var=%s", ftbHost, datasetName, d.Name),
					ID:  d.Options[0],
				},
			}
		}
	}

	doc := &models.ObservationsDoc{
		Dimensions:        dimensionsMap,
		Limit:             0,
		Links:             nil,
		Observations:      observations,
		Offset:            0,
		TotalObservations: 0,
		UnitOfMeasure:     "",
		UsageNotes:        nil,
	}

	return doc, nil
}
