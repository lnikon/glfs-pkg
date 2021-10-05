package server

import (
	"fmt"

	glkube "github.com/lnikon/glfs-pkg/pkg/kube"
)

type Computation struct {
	algorithm Algorithm
}

type ComputationService struct {
}

func (c *ComputationService) GetAllComputations() []Computation {
	return []Computation{{algorithm: Kruskal}}
}

func (c *ComputationService) PostComputation(request *PostComputationRequest) PostComputationResponse {
	fmt.Printf("Post computation request called for %s\n", request.Algorithm)
	glkube.CreateDeployment()
	return PostComputationResponse{}
}
