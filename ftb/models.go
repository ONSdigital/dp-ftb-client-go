package ftb

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

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

type RulesError struct {
	msg string
}

func (fq *FilterQuery) newQueryRequest(host, authToken string) (*http.Request, error) {
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

func (r *QueryResult) IsBlockedByRules() bool {
	return r.EvalCatOffsetLenPairs != nil && len(r.EvalCatOffsetLenPairs) > 0
}

func (r *QueryResult) getBlockedCodeIndices() ([]int, error) {
	if len(r.EvalCatOffsetLenPairs)%2 != 0 {
		return nil, errors.New("incorrect input")
	}

	codes := make([]int, 0)
	for i := 0; i < len(r.EvalCatOffsetLenPairs); i += 2 {
		startIndex := r.EvalCatOffsetLenPairs[i]
		count := r.EvalCatOffsetLenPairs[i+1]

		endIndex := startIndex + count - 1

		for i := startIndex; i <= endIndex; i++ {
			codes = append(codes, i)
		}
	}

	return codes, nil
}

func newRulesError(dimension string, codes []string) error {
	return RulesError{
		msg: fmt.Sprintf("filter unsuccessful disclosure control applied to Dimension codes : %s, [%s]", dimension, strings.Join(codes, ",")),
	}
}

func (err RulesError) Error() string {
	return err.msg
}
