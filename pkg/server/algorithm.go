package server

import (
    glconstants "github.com/lnikon/glfs-pkg/pkg/constants"
)

type AlgorithmService struct {
}

func NewAlgorithmService() *AlgorithmService {
	return &AlgorithmService{}
}

func (a *AlgorithmService) Algorithm() []glconstants.Algorithm {
	return []glconstants.Algorithm{glconstants.Kruskal, glconstants.Prim}
}
