package ftb

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func newQueryRequest(fq Query, host, authToken string) (*http.Request, error) {
	ftbURL, err := url.Parse(fmt.Sprintf("%s/v6/query/%s?", host, fq.DatasetName))
	if err != nil {
		return nil, err
	}

	q := ftbURL.Query()
	for _, d := range fq.DimensionsOptions {
		dimName := strings.ToUpper(d.Name)
		q.Add(dimParam, dimName)

		if len(d.Options) > 0 {
			options := make([]string,0)
			for _, v := range d.Options {
				if v != "" {
					options = append(options, v)
				}
			}

			if len(options) > 0 {
				q.Add(includeParam, fmt.Sprintf("%s,%s", dimName, strings.Join(options, ",")))
			}
		}
	}

	q.Add("limit", strconv.Itoa(fq.Limit))

	ftbURL.RawQuery = q.Encode()
	r, err := httpRequestWithAuthHeader(authToken, http.MethodGet, ftbURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func newGetDimensionByIndexRequest(host, authToken, dataset, dimension string, index int) (*http.Request, error) {
	reqURL := fmt.Sprintf("%s/v6/datasets/%s/dimensions/%s/index/%d", host, dataset, strings.ToUpper(dimension), index)
	return httpRequestWithAuthHeader(authToken, http.MethodGet, reqURL, nil)
}

func httpRequestWithAuthHeader(authToken, method, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Authorization", "Bearer "+authToken)
	return r, nil
}

func newGetDimensionReq(host, authToken, dataset, dimension string) (*http.Request, error) {
	ftbURL := fmt.Sprintf("%s/v6/codebook/%s?var=%s", host, dataset, dimension)
	return httpRequestWithAuthHeader(authToken, http.MethodGet, ftbURL, nil)
}
