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
)

type Client struct {
	AuthToken string
	Host      string
	HttpCli   dphttp.Clienter
}

func (c *Client) Query(ctx context.Context, q FilterQuery) error {
	r, err := q.newQueryRequest(c.Host, c.AuthToken)
	if err != nil {
		return err
	}

	result, err := c.doQuery(ctx, r)
	if err != nil {
		return err
	}

	if result.IsBlockedByRules() {
		return c.blockedByRulesError(ctx, q.DatasetName, q.RootDimension, result)
	}

	return nil
}

func (c *Client) GetDimensionByIndex(ctx context.Context, dataset, dimension string, index int) (*DimensionResponse, error) {
	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v6/datasets/%s/dimensions/%s/index/%d", c.Host, dataset, dimension, index), nil)
	if err != nil {
		return nil, err
	}

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

	var dim DimensionResponse
	err = json.Unmarshal(body, &dim)
	if err != nil {
		return nil, err
	}

	return &dim, nil
}

func (c *Client) doQuery(ctx context.Context, r *http.Request) (*QueryResult, error) {
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

	var result QueryResult
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

func (c *Client) blockedByRulesError(ctx context.Context, dataset, rootDimension string, res *QueryResult) error {
	blockedCodeIndices, err := res.getBlockedCodeIndices()
	if err != nil {
		return err
	}

	blockedDimensionCodes, err := c.getCodesBlockedByRules(ctx, dataset, rootDimension, blockedCodeIndices)
	if err != nil {
		return err
	}

	return newRulesError(rootDimension, blockedDimensionCodes)
}

func (c *Client) getCodesBlockedByRules(ctx context.Context, dataset, dimension string, indices []int) ([]string, error) {
	codes := make([]string, 0)

	for _, index := range indices {
		dim, err := c.GetDimensionByIndex(ctx, dataset, dimension, index)
		if err != nil {
			return nil, err
		}
		codes = append(codes, dim.Name)
	}
	return codes, nil
}
