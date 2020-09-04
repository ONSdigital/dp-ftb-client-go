package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/fatih/color"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run() error {
	cli := dphttp.DefaultClient

	requests := map[string]string{
		"Single Observation: [Country=Wales, Age=31, Sex=Male]":                 "http://localhost:24500/datasets/People/editions/time-series/versions/1/observations?country=synE92000001&age=31&sex=1",
		"Wildcard 1 dimension: [Country=Wales, Ages=31, Sex=Male,Female]":       "http://localhost:24500/datasets/People/editions/time-series/versions/1/observations?country=synE92000001&age=31&sex",
		"Dimension with multiple options: [Country=Wales, Age=30,31, Sex=Male]": "http://localhost:24500/datasets/People/editions/time-series/versions/1/observations?country=synE92000001&age=30&age=31&sex=1",
		"Multiple wildcards: [Country=Wales, Age=30,31, Sex=Male]":              "http://localhost:24500/datasets/People/editions/time-series/versions/1/observations?country=synE92000001&age&sex",
	}

	for desc, url := range requests {
		err := exec(cli, desc, url)
		if err != nil {
			return err
		}

		<-time.After(time.Second * 3)
	}

	return nil
}

func exec(cli dphttp.Clienter, desc, url string) error {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := cli.Do(context.Background(), r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(string(b))
	}

	var entity interface{}
	err = json.Unmarshal(b, &entity)
	if err != nil {
		return err
	}

	body, err := json.MarshalIndent(entity, "", "  ")
	if err != nil {
		return err
	}


	color.New(color.FgGreen, color.Bold).Printf("\n\n%s\n", desc)
	color.New(color.FgCyan, color.Bold).Printf("\n%s\n", body)
	return nil
}
