package ftb

import (
	"errors"
	"os"
	"strconv"

	"github.com/ONSdigital/dp-ftb-client-go/codebook"
	"github.com/olekukonko/tablewriter"
)

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

func getObservationsTable(datasetName string, queryOptions []DimensionOptions, dimensions map[string]*codebook.Dimension, observations []int) (*ObservationsTable, error) {
	table, err := newEmptyObservationsTable(datasetName, queryOptions, dimensions)
	if err != nil {
		return nil, err
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

// Create a new table by calculating all permutations of the dimension options provided in order
func newEmptyObservationsTable(datasetName string, queryOptions []DimensionOptions, dimensions map[string]*codebook.Dimension) (*ObservationsTable, error) {
	header := make([]string, 0)
	for _, d := range queryOptions {
		header = append(header, d.Name)
	}

	header = append(header, "Observation")

	t := &ObservationsTable{
		Header: header,
		Rows:   make([][]string, 0),
	}

	for _, dim := range queryOptions {
		details := dimensions[dim.Name]
		options := dim.Options

		if len(t.Rows) == 0 {
			for _, opt := range options {
				label := details.GetLabelByCode(opt)
				t.Rows = append(t.Rows, []string{label})
			}
			continue
		}

		// Cant iterate and update an array at the same time.
		// Create a new array to update and then assign to the original after the iteration.
		update := make([][]string, 0)

		for _, currentValue := range t.Rows {
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

		t.Rows = update
	}

	return t, nil
}
