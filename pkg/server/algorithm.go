package server

type Algorithm string

const (
	Kruskal = "kruskal"
	Prim    = "mst"
)

type AlgorithmService struct {
}

func NewAlgorithmService() *AlgorithmService {
	return &AlgorithmService{}
}

func (a *AlgorithmService) Algorithm() []Algorithm {
	return []Algorithm{Kruskal, Prim}
}
