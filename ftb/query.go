package ftb

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/log.go/log"
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
	Limit             int
}

type DimensionOptions struct {
	Name    string
	Options []string
}

// return
type QueryResult struct {
	DisclosureControlDetails *DisclosureControlDetails `json:"disclosure_control_details,omitempty"`
	V4Table                  *V4Table                  `json:"observations,omitempty"`
}

// return
type DisclosureControlDetails struct {
	Status         string   `bson:"status"          json:"status,omitempty"`
	Dimension      string   `bson:"dimension"       json:"dimension,omitempty"`
	BlockedOptions []string `bson:"blocked_options" json:"blocked_options,omitempty"`
	BlockedCount   int      `bson:"blocked_count" json:"blocked_count,omitempty"`
}

func NewQuery(dataset, rootDim string, params map[string][]string) Query {
	q := Query{
		DatasetName:       dataset,
		RootDimension:     rootDim,
		DimensionsOptions: make([]DimensionOptions, 0),
	}

	for k, v := range params {
		options := make([]string, 0)
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
		V4Table:                  nil,
	}

	if dcStatus.Status == StatusBlocked {
		return result, nil
	}

	if q.Limit > 0 {
		dimensions, err := c.getDimensionDetails(q.DatasetName, q.DimensionsOptions)
		if err != nil {
			return nil, err
		}
		log.Event(ctx, "getDimensionDetails completed", log.INFO)

		table, err := getAsV4Table(q.DatasetName, q.DimensionsOptions, dimensions, resp.Counts)
		if err != nil {
			return nil, err
		}
		log.Event(ctx, "getAsV4Table completed", log.INFO)

		result.V4Table = table
	}
	return result, nil
}

func (c *client) doQuery(ctx context.Context, r *http.Request) (*queryResponse, error) {
	logD := log.Data{"url": r.URL.String()}
	log.Event(ctx, "executing FTB query request", logD, log.INFO)

	resp, err := c.HttpCli.Do(ctx, r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logD["status"] = resp.StatusCode
		logD["response_body"] = string(body)
		log.Event(ctx, "FTB query request returned failure status code", logD, log.INFO)
		return nil, fmt.Errorf("ftb status error: %d", resp.StatusCode)
	}

	var result queryResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	log.Event(ctx, "FTB query request completed successfully", logD, log.INFO)
	return &result, nil
}

func (c *client) getDCStatus(ctx context.Context, resp *queryResponse, datasetName, rootDimension string) (*DisclosureControlDetails, error) {
	if resp.BlockedByRules() {
		blockedCodeIndices, err := resp.getBlockedCodeIndices()
		if err != nil {
			return nil, err
		}

		return &DisclosureControlDetails{
			Status:       StatusBlocked,
			Dimension:    rootDimension,
			BlockedCount: len(blockedCodeIndices),
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
