package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ONSdigital/dp-ftb-client-go/ftb"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

func main() {
	err := run()
	if err != nil {
		log.Event(nil, "borked", log.Error(err), log.ERROR)
		os.Exit(1)
	}
}

func run() error {
	ftbCli := ftb.NewClient(fmt.Sprintf("http://%s:10100", os.Getenv("EC2_IP")), os.Getenv("AUTH_PROXY_TOKEN"), dphttp.DefaultClient)

	q := ftb.Query{
		DatasetName: "People",
		DimensionsOptions: []ftb.DimensionOptions{
			{Name: "COUNTRY", Options: []string{"synE92000001", "synW92000004"}},
			{Name: "AGE_2CATS", Options: []string{"0-15", "16-90"}},
			//{Name: "AGE", Options: []string{"20", "21", "22", "23", "34"}},
			{Name: "SEX", Options: []string{"1", "2"}},
		},
		RootDimension: "COUNTRY",
	}

	result, err := ftbCli.Query(context.Background(), q)
	if err != nil {
		return err
	}

	if result.IsBlocked() {
		fmt.Println("query restricted by disclosure controls")
		return nil
	}

	result.ObservationsTable.Print()
	return nil
}
