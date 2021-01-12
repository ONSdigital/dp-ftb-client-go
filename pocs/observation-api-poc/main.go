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

// For the purposes of the POC it easier to create a new type embedding the models.ObservationsDoc and add a new DisclosureControlDetails field.
type ExtendedObservationsResponse struct {
	*models.ObservationsDoc
	DisclosureControlDetails *ftb.DisclosureControlDetails
}

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

func getObservations(ctx context.Context, datasetName string, queryParams url.Values) (*ExtendedObservationsResponse, error) {
	// For POC purposes have to manually set the value. In the real use case we will need to be able to determine this at run time.
	query := ftb.NewQuery(datasetName, "OA", queryParams) //synW00007765
	result, err := ftbCli.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	observationsDoc := &ExtendedObservationsResponse{
		ObservationsDoc: &models.ObservationsDoc{},
		DisclosureControlDetails: &ftb.DisclosureControlDetails{
			Status:         ftb.StatusOK,
			Dimension:      "",
			BlockedOptions: nil,
			BlockedCount:   0,
		},
	}

	if result.IsBlocked() {
		observationsDoc.DisclosureControlDetails = result.DisclosureControlDetails
		return observationsDoc, nil
	}

	log.Event(ctx, "query completed", log.INFO)

	wildCards := make(map[string]ftb.DimensionOptions, 0)
	multiSelection := make(map[string]ftb.DimensionOptions, 0)

	for _, d := range query.DimensionsOptions {
		if len(d.Options) == 0 {
			wildCards[d.Name] = d
		} else if len(d.Options) > 1 {
			multiSelection[d.Name] = d
		}
	}

	observations := make([]models.Observation, 0)

	for _, row := range result.V4Table.Rows {
		dims := make(map[string]*models.DimensionObject, 0)

		for i := 0; i < len(row)-1; i += 2 {
			dimName := result.V4Table.Header[i]

			if isWildcardOrMultiSelection(dimName, wildCards, multiSelection) {
				dims[result.V4Table.Header[i]] = &models.DimensionObject{
					HRef:  fmt.Sprintf("%s/v8/codebook/%s?var=%s", ftbHost, datasetName, dimName),
					ID:    row[i+1],
					Label: row[i],
				}
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
					URL: getLink(datasetName, d.Name),
					ID:  d.Options[0],
				},
			}
		}
	}

	observationsDoc.Dimensions = dimensionsMap
	observationsDoc.Observations = observations
	observationsDoc.TotalObservations = len(observations)

	return observationsDoc, nil
}

func getLink(datasetName, dimensionName string) string {
	return fmt.Sprintf("%s/v8/codebook/%s?var=%s", ftbHost, datasetName, dimensionName)
}

func isWildcardOrMultiSelection(dimensionName string, wildCards map[string]ftb.DimensionOptions, multiSelections map[string]ftb.DimensionOptions) bool {
	_, isWildCard := wildCards[dimensionName]
	_, isMultiSelection := multiSelections[dimensionName]

	return isWildCard || isMultiSelection
}
