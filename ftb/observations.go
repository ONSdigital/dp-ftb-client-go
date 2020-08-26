package ftb

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/ONSdigital/dp-ftb-client-go/codebook"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/olekukonko/tablewriter"
)

var ftbCli = NewClient("", os.Getenv("AUTH_PROXY_TOKEN"), dphttp.DefaultClient)

type ObservationsTable struct {
	Header []string
	Rows   [][]string
}

func (o *ObservationsTable) Print() {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(o.Header)

	for _, r := range o.Rows {
		tw.Append(r)
	}

	tw.Render()
}

func createObservationsTable(datasetName string, dimensionOptions []DimensionOptions, observations []int) (*ObservationsTable, error) {
	table := newObservationsTable(dimensionOptions)

	dimensionDetailsMap, err := getDimensionDetails(datasetName, dimensionOptions)
	if err != nil {
		return nil, err
	}

	for _, dim := range dimensionOptions {
		details := dimensionDetailsMap[dim.Name]
		options := dim.Options

		if len(table.Rows) == 0 {
			for _, opt := range options {
				label := details.GetLabelByCode(opt)
				table.Rows = append(table.Rows, []string{label})
			}
			continue
		}

		// Cant iterate and update an array at the same time.
		// Create a new array to update and then assign to the original after the iteration.
		update := make([][]string, 0)

		for _, currentValue := range table.Rows {
			for _, opt := range options {

				// copy the current row value
				newRow := make([]string, 0)
				newRow = append(newRow, currentValue...)

				// append the new value the row
				label := details.GetLabelByCode(opt)
				newRow = append(newRow, label)

				// update the tracking copy.
				update = append(update, newRow)
			}
		}

		table.Rows = update
	}

	if len(table.Rows) != len(observations) {
		return nil, errors.New("BORK")
	}

	for i, count := range observations {
		r := table.Rows[i]
		r = append(r, strconv.Itoa(count))
		table.Rows[i] = r
	}

	return table, nil
}

func getDimensionDetails(dataset string, dims []DimensionOptions) (map[string]*codebook.Dimension, error) {
	mapping := make(map[string]*codebook.Dimension, 0)

	for _, d := range dims {
		details, err := ftbCli.GetDimension(context.Background(), dataset, d.Name)
		if err != nil {
			return nil, err
		}

		mapping[d.Name] = details
	}

	return mapping, nil
}

func newObservationsTable(dims []DimensionOptions) *ObservationsTable {
	header := make([]string, 0)
	for _, d := range dims {
		header = append(header, d.Name)
	}

	header = append(header, "Observation")

	return &ObservationsTable{
		Header: header,
		Rows:   make([][]string, 0),
	}
}
