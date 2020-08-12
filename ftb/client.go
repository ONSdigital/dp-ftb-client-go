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

	resp, err := c.HttpCli.Do(context.Background(), r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ftb status error: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result QueryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	log.Event(ctx, "FTB query response", log.INFO, log.Data{
		"url":    r.URL.String(),
		"result": result,
	})

	return nil
}
