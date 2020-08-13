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

func (c *Client) Query(ctx context.Context, query FilterQuery) error {
	r, err := query.createRequest(c.Host, c.AuthToken)
	if err != nil {
		return err
	}

	result, err := c.doQuery(ctx, r)
	if err != nil {
		return err
	}

	if result.IsBlockedByDisclosureControl() {
		return c.newDisclosureControlErr(ctx, query.DatasetName, query.RootDimension, result.EvalCatOffsetLenPairs)
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

func (c *Client) newDisclosureControlErr(ctx context.Context, dataset, dimension string, blocked []int) error {
	codes := make([]string, 0)

	for i := 0; i < len(blocked); i += 2 {
		start := blocked[i]
		end := start + (blocked[i+1] - 1)

		for i := start; i <= end; i++ {
			dim, err := c.GetDimensionByIndex(ctx, dataset, dimension, i)
			if err != nil {
				return err
			}
			codes = append(codes, dim.Name)
		}

	}
	return DisclosureControlError{
		dimension: dimension,
		codes:     codes,
	}
}
