package ftb

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/ONSdigital/dp-ftb-client-go/codebook"
)

func (c *client) GetDimension(ctx context.Context, dataset, dimension string) (*codebook.Dimension, error) {
	req, err := newGetDimensionReq(c.Host, c.AuthToken, dataset, dimension)
	if err != nil {
		return nil, err
	}

	resp, err := c.HttpCli.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("incorrect status code expected 200 but was %d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cb codebook.Codebook
	err = json.Unmarshal(b, &cb)
	if err != nil {
		return nil, err
	}

	return &cb.CodeBook[0], nil
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

func (c *client) getDimensionDetails(dataset string, dims []DimensionOptions) (map[string]*codebook.Dimension, error) {
	mapping := make(map[string]*codebook.Dimension, 0)

	for _, d := range dims {
		details, err := c.GetDimension(context.Background(), dataset, d.Name)
		if err != nil {
			return nil, err
		}

		mapping[d.Name] = details
	}

	return mapping, nil
}
