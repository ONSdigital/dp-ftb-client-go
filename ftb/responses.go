package ftb

import "errors"

type GetDimensionOptionResponse struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	Code  string `json:"code"`
}

type queryResponse struct {
	Counts                []int              `json:"counts"`
	DatasetDigest         string             `json:"datasetDigest"`
	Dimensions            []dimensionDetails `json:"dimensions"`
	EvalCatOffsetLenPairs []int              `json:"evalCatOffsetLenPairs"`
}

type dimensionDetails struct {
	Name              string `json:"name"`
	CatOffsetLenPairs []int  `json:"catOffsetLenPairs"`
}

func (r *queryResponse) BlockedByRules() bool {
	return r.EvalCatOffsetLenPairs != nil && len(r.EvalCatOffsetLenPairs) > 0
}

func (r *queryResponse) getBlockedCodeIndices() ([]int, error) {
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
