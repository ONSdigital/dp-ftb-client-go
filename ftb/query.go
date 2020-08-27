package ftb

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	dimParam     = "dim"
	includeParam = "incl"

	StatusOK      = "OK"
	StatusBlocked = "Blocked"
)

// input
type Query struct {
	DatasetName       string
	DimensionsOptions []DimensionOptions
	RootDimension     string
}

type DimensionOptions struct {
	Name    string
	Options []string
}

// return
type QueryResult struct {
	DisclosureControlDetails *DisclosureControlDetails `json:"disclosure_control_details,omitempty"`
	ObservationsTable        *ObservationsTable        `json:"observations,omitempty"`
}

// return
type DisclosureControlDetails struct {
	Status         string   `bson:"status"          json:"status,omitempty"`
	Dimension      string   `bson:"dimension"       json:"dimension,omitempty"`
	BlockedOptions []string `bson:"blocked_options" json:"blocked_options,omitempty"`
}

func NewQuery(dataset, rootDim string, params map[string][]string) Query {
	q := Query{
		DatasetName:       dataset,
		RootDimension:     rootDim,
		DimensionsOptions: make([]DimensionOptions, 0),
	}

	for k, v := range params {
		options := make([]string,0)
		for _, opt := range v {
			if opt != "" {
				options = append(options, opt)
			}
		}

		q.DimensionsOptions = append(q.DimensionsOptions, DimensionOptions{Name: k, Options: options})
	}

	return q
}

func (c *client) Query(ctx context.Context, q Query) (*QueryResult, error) {
	r, err := newQueryRequest(q, c.Host, c.AuthToken)
	if err != nil {
		return nil, err
	}

	resp, err := c.doQuery(ctx, r)
	if err != nil {
		return nil, err
	}

	dcStatus, err := c.getDCStatus(ctx, resp, q.DatasetName, q.RootDimension)
	if err != nil {
		return nil, err
	}

	result := &QueryResult{
		DisclosureControlDetails: dcStatus,
		ObservationsTable:        nil,
	}

	if dcStatus.Status == StatusBlocked {
		return result, nil
	}

	dimensions, err := c.getDimensionDetails(q.DatasetName, q.DimensionsOptions)
	if err != nil {
		return nil, err
	}

	table, err := getObservationsTable(q.DatasetName, q.DimensionsOptions, dimensions, resp.Counts)
	if err != nil {
		return nil, err
	}

	result.ObservationsTable = table
	return result, nil
}

func (c *client) doQuery(ctx context.Context, r *http.Request) (*queryResponse, error) {
	resp, err := c.HttpCli.Do(ctx, r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ftb status error: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result queryResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) getDCStatus(ctx context.Context, resp *queryResponse, datasetName, rootDimension string) (*DisclosureControlDetails, error) {
	if resp.BlockedByRules() {
		blockedCodeIndices, err := resp.getBlockedCodeIndices()
		if err != nil {
			return nil, err
		}

		blockedDimensionCodes := make([]string, 0)

		for _, index := range blockedCodeIndices {
			dim, err := c.GetDimensionByIndex(ctx, datasetName, rootDimension, index)
			if err != nil {
				return nil, err
			}

			blockedDimensionCodes = append(blockedDimensionCodes, dim.Code)
		}

		return &DisclosureControlDetails{
			Status:         StatusBlocked,
			Dimension:      rootDimension,
			BlockedOptions: blockedDimensionCodes,
		}, nil
	}

	return &DisclosureControlDetails{
		Status:         StatusOK,
		Dimension:      rootDimension,
		BlockedOptions: []string{},
	}, nil
}


func (r *QueryResult) IsBlocked() bool {
	return r.DisclosureControlDetails.Status == StatusBlocked
}
