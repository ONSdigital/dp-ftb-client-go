package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	dphttp "github.com/ONSdigital/dp-net/http"
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

func (c *Client) Query(dims []Dimension) error {
	r, err := c.createQueryReq(dims)
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

	var blob map[string]interface{}
	err = json.Unmarshal(body, &blob)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(blob, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("FTB query response\n%s\n", string(b))
	return nil
}

func (c *Client) createQueryReq(dimensions []Dimension) (*http.Request, error) {
	ftbURL, err := url.Parse(fmt.Sprintf("%s/v6/query/People?", c.Host))
	if err != nil {
		return nil, err
	}

	q := ftbURL.Query()
	for _, dim := range dimensions {
		if len(dim.GetOptions()) > 0 {
			dimName := strings.ToUpper(dim.GetName())
			q.Add(dimParam, dimName)
			q.Add(includeParam, GetIncludeParam(dim))
		}
	}

	ftbURL.RawQuery = q.Encode()
	r, err := http.NewRequest(http.MethodGet, ftbURL.String(), nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Authorization", "Bearer "+c.AuthToken)
	return r, nil
}

func GetIncludeParam(d Dimension) string {
	return fmt.Sprintf("%s,%s", d.GetName(), strings.Join(d.GetOptions(), ","))
}
