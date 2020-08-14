package main

import (
	"context"
	"os"

	"github.com/ONSdigital/dp-ftb-client-go/ftb"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

func main() {
	if err := run(); err != nil {
		log.Event(nil, "whoops", log.Error(err), log.ERROR)
		os.Exit(1)
	}
}

func run() error {
	ftbcli := ftb.NewClient("http://localhost:10100", os.Getenv("AUTH_PROXY_TOKEN"), dphttp.DefaultClient)

	query := ftb.Query{
		DatasetName:   "People",
		RootDimension: "OA",
		DimensionsOptions: []ftb.DimensionOptions{
			{Name: "OA", Options: []string{"synW00000005"}},
			{Name: "Age", Options: []string{"30"}},
			{Name: "Sex", Options: []string{"1"}},
		},
	}

	result, err := ftbcli.Query(context.Background(), query)
	if err != nil {
		return err
	}

	log.Event(context.Background(), "query response", log.INFO, log.Data{"response": result})
	return nil
}
