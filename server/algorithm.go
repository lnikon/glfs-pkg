package main

type Algorithm string

const (
	Kruskal = "kruskal"
	Prim    = "mst"
)

type AlgorithmService struct {
}

func (a *AlgorithmService) Algorithm() []Algorithm {
	return []Algorithm{Kruskal, Prim}
}
