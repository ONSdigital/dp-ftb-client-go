package ftb

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

const (
	dimParam     = "dim"
	includeParam = "incl"

	StatusOK      = "OK"
	StatusBlocked = "Blocked"
)

type Clienter interface {
	Query(ctx context.Context, q Query) (*QueryResult, error)
	GetDimensionByIndex(ctx context.Context, dataset, dimension string, index int) (*GetDimensionOptionResponse, error)
}

type client struct {
	AuthToken string
	Host      string
	HttpCli   dphttp.Clienter
}

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
	Status                   string
	DisclosureControlDetails *DisclosureControlDetails
}

// return
type DisclosureControlDetails struct {
	Dimension      string   `bson:"dimension"       json:"dimension,omitempty"`
	BlockedOptions []string `bson:"blocked_options" json:"blocked_options,omitempty"`
}

func NewClient(host, authToken string, httpCli dphttp.Clienter) Clienter {
	return &client{
		AuthToken: authToken,
		Host:      host,
		HttpCli:   httpCli,
	}
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

	return c.mapResponseToFilterResult(ctx, resp, q.DatasetName, q.RootDimension)
}

func (c *client) mapResponseToFilterResult(ctx context.Context, resp *queryResponse, dataset, rootDim string, ) (*QueryResult, error) {
	if !resp.BlockedByRules() {
		return &QueryResult{Status: StatusOK, DisclosureControlDetails: nil}, nil
	}

	blockedCodeIndices, err := resp.getBlockedCodeIndices()
	if err != nil {
		return nil, err
	}

	blockedDimensionCodes := make([]string, 0)

	for _, index := range blockedCodeIndices {
		dim, err := c.GetDimensionByIndex(ctx, dataset, rootDim, index)
		if err != nil {
			return nil, err
		}

		blockedDimensionCodes = append(blockedDimensionCodes, dim.Code)
	}

	result := &QueryResult{
		Status: StatusBlocked,
		DisclosureControlDetails: &DisclosureControlDetails{
			Dimension:      rootDim,
			BlockedOptions: blockedDimensionCodes,
		},
	}

	return result, nil
}

func (c *client) GetDimensionByIndex(ctx context.Context, dataset, dimension string, index int) (*GetDimensionOptionResponse, error) {
	r, err := newGetDimensionByIndexRequest(c.Host, c.AuthToken, dataset, dimension, index)

	resp, err := c.HttpCli.Do(ctx, r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var dim GetDimensionOptionResponse
	err = json.Unmarshal(body, &dim)
	if err != nil {
		return nil, err
	}

	return &dim, nil
}

func (c *client) doQuery(ctx context.Context, r *http.Request) (*queryResponse, error) {
	resp, err := c.HttpCli.Do(context.Background(), r)
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

	log.Event(ctx, "FTB query response", log.INFO, log.Data{
		"url":    r.URL.String(),
		"result": result,
	})

	return &result, nil
}

func (q *QueryResult) IsBlockedByRules() bool {
	return q.Status == StatusBlocked
}