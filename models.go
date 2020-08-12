package client

type Dimension interface {
	GetName() string
	GetOptions() []string
}
