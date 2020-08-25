package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ONSdigital/dp-ftb-client-go/ftb"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

var (
	// in the real world this would be a codebook look up.
	codebookLabels = map[string]string{
		"synE92000001": "England",
		"synW92000004": "Wales",
		"0-15":         "Age 0 to 15",
		"16-90":        "Age 16 and over",
		"1":            "Male",
		"2":            "Female",
	}
)

func main() {
	if err := run(); err != nil {
		log.Event(nil, "whoops", log.Error(err), log.ERROR)
		os.Exit(1)
	}
}

func run() error {
	q := ftb.Query{
		DatasetName: "People",
		DimensionsOptions: []ftb.DimensionOptions{
			{Name: "COUNTRY", Options: []string{"synE92000001", "synW92000004"}},
			{Name: "AGE_2CATS", Options: []string{"0-15", "16-90"}},
			{Name: "SEX", Options: []string{"1", "2"}},
		},
		RootDimension: "COUNTRY",
	}

	ftbCli := ftb.NewClient("http://99.80.12.125:10100", os.Getenv("AUTH_PROXY_TOKEN"), dphttp.DefaultClient)

	res, err := ftbCli.Query(context.Background(), q)
	if err != nil {
		return err
	}

	rows := getObservationPermutations(q)
	observationValues := res.Counts

	if len(rows) != len(observationValues) {
		return errors.New("BORK")
	}

	fmt.Println()
	for i, r := range rows {
		fmt.Printf("\t%s %d\n", r, observationValues[i])
	}

	return nil
}

func getObservationPermutations(query ftb.Query) []string {
	permutations := make([]string, 0)

	for _, dim := range query.DimensionsOptions {
		options := dim.Options

		if len(permutations) == 0 {
			for _, opt := range options {
				permutations = append(permutations, codebookLabels[opt])
			}
			continue
		}

		updated := make([]string, 0)

		for _, currentValue := range permutations {
			for _, opt := range options {
				label := codebookLabels[opt]
				updated = append(updated, fmt.Sprintf("%s %s", currentValue, label))
			}
		}

		permutations = updated
	}

	return permutations
}
