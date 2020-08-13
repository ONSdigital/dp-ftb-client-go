package ftb

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type DisclosureControlError struct {
	dimension string
	codes     []string
}

type DimensionDetails struct {
	Name  string
	Code  string
	Index int
}

type DimensionResponse struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	Code  string `json:"code"`
}

type FilterQuery struct {
	DatasetName   string
	Dimensions    []QueryDimension
	RootDimension string
}

type QueryDimension struct {
	Name    string
	Options []string
}

type QueryResult struct {
	Counts                []int       `json:"counts"`
	DatasetDigest         string      `json:"datasetDigest"`
	Dimensions            []Dimension `json:"dimensions"`
	EvalCatOffsetLenPairs []int       `json:"evalCatOffsetLenPairs"`
}

type Dimension struct {
	Name              string `json:"name"`
	CatOffsetLenPairs []int  `json:"catOffsetLenPairs"`
}

func (fq *FilterQuery) createRequest(host, authToken string) (*http.Request, error) {
	ftbURL, err := url.Parse(fmt.Sprintf("%s/v6/query/%s?", host, fq.DatasetName))
	if err != nil {
		return nil, err
	}

	q := ftbURL.Query()
	for _, d := range fq.Dimensions {
		if len(d.Options) > 0 {
			dimName := strings.ToUpper(d.Name)
			q.Add(dimParam, dimName)
			q.Add(includeParam, fmt.Sprintf("%s,%s", dimName, strings.Join(d.Options, ",")))
		}
	}

	ftbURL.RawQuery = q.Encode()
	r, err := http.NewRequest(http.MethodGet, ftbURL.String(), nil)
	if err != nil {
		return nil, err
	}

	if len(authToken) > 0 {
		r.Header.Set("Authorization", "Bearer "+authToken)
	}
	return r, nil
}

func (r *QueryResult) IsBlockedByDisclosureControl() bool {
	return r.EvalCatOffsetLenPairs != nil && len(r.EvalCatOffsetLenPairs) > 0
}

func (err DisclosureControlError) Error() string {
	return fmt.Sprintf("Disclosure control applied to Dimension: %s, [%s]", err.dimension, strings.Join(err.codes, ","))
}
