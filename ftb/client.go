package ftb

import (
	"context"

	"github.com/ONSdigital/dp-ftb-client-go/codebook"
	dphttp "github.com/ONSdigital/dp-net/http"
)

type Clienter interface {
	Query(ctx context.Context, q Query) (*QueryResult, error)
	GetDimension(ctx context.Context, dataset, dimension string) (*codebook.Dimension, error)
	GetDimensionByIndex(ctx context.Context, dataset, dimension string, index int) (*GetDimensionOptionResponse, error)
}

type client struct {
	AuthToken string
	Host      string
	HttpCli   dphttp.Clienter
}

func NewClient(host, authToken string, httpCli dphttp.Clienter) Clienter {
	return &client{
		AuthToken: authToken,
		Host:      host,
		HttpCli:   httpCli,
	}
}
