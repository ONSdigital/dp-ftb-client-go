package ftb

import (
	"errors"
	"os"
	"strconv"

	"github.com/ONSdigital/dp-ftb-client-go/codebook"
	"github.com/olekukonko/tablewriter"
)

type V4Table struct {
	Header []string
	Rows   [][]string
}

func (o *V4Table) Print() {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(o.Header)

	for _, r := range o.Rows {
		tw.Append(r)
	}

	tw.Render()
}

func getAsV4Table(datasetName string, queryOptions []DimensionOptions, dimensions map[string]*codebook.Dimension, observations []int) (*V4Table, error) {
	table, err := newEmptyV4Table(datasetName, queryOptions, dimensions)
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
func newEmptyV4Table(datasetName string, queryOptions []DimensionOptions, dimensions map[string]*codebook.Dimension) (*V4Table, error) {
	header := make([]string, 0)
	for _, d := range queryOptions {
		header = append(header, d.Name, d.Name + " code")
	}

	header = append(header, "Observation")

	t := &V4Table{
		Header: header,
		Rows:   make([][]string, 0),
	}

	for _, dim := range queryOptions {
		details := dimensions[dim.Name]
		if len(dim.Options) == 0 {
			dim.Options = append(dim.Options, details.Codes...)
		}

		if len(t.Rows) == 0 {
			for _, opt := range dim.Options {
				label := details.GetLabelByCode(opt)
				t.Rows = append(t.Rows, []string{label, opt})
			}
			continue
		}

		// Cant iterate and update an array at the same time.
		// Create a new array to update and then assign to the original after the iteration.
		update := make([][]string, 0)

		for _, currentValue := range t.Rows {
			for _, opt := range dim.Options {

				// copy the current row value
				newRow := make([]string, 0)
				newRow = append(newRow, currentValue...)

				// append the new value the row
				label := details.GetLabelByCode(opt)
				newRow = append(newRow, []string{label, opt}...)

				// update the tracking copy.
				update = append(update, newRow)
			}
		}

		t.Rows = update
	}

	return t, nil
}
