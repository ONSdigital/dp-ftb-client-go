package ftb

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func newQueryRequest(fq Query, host, authToken string) (*http.Request, error) {
	ftbURL, err := url.Parse(fmt.Sprintf("%s/v6/query/%s?", host, fq.DatasetName))
	if err != nil {
		return nil, err
	}

	q := ftbURL.Query()
	for _, d := range fq.DimensionsOptions {
		if len(d.Options) > 0 {
			dimName := strings.ToUpper(d.Name)
			q.Add(dimParam, dimName)
			q.Add(includeParam, fmt.Sprintf("%s,%s", dimName, strings.Join(d.Options, ",")))
		}
	}

	ftbURL.RawQuery = q.Encode()
	r, err := httpRequestWithAuthHeader(authToken, http.MethodGet, ftbURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func newGetDimensionByIndexRequest(host, authToken, dataset, dimension string, index int) (*http.Request, error) {
	reqURL := fmt.Sprintf("%s/v6/datasets/%s/dimensions/%s/index/%d", host, dataset, dimension, index)
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
