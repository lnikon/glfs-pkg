package server

import (
	"fmt"

	glkube "github.com/lnikon/glfs-pkg/pkg/kube"
	glconstants "github.com/lnikon/glfs-pkg/pkg/constants"
)

const (
	ComputationDeploymentNamePattern = "computation-%d"
)

type Computation struct {
	Algorithm glconstants.Algorithm
	Name      string
}

type ComputationService struct {
	computations []Computation
}

func NewComputationService() (*ComputationService, error) {
	computationService := &ComputationService{}

	// deploymentsList := glkube.GetAllDeployments()
	// if deploymentsList == nil {
	// 	return nil, errors.New("unable to get all deployments")
	// }

	// computationDeploymentNameRegexp, err := regexp.Compile("(computation-[0-9]+)")
	// if err != nil {
	// 	log.Fatal("Unable to compile regexp")
	// 	return nil, errors.New("unable to compile computation name matching regexp")
	// }

	// for _, deployment := range deploymentsList.Items {
	// 	name := deployment.ObjectMeta.Name
	// 	if computationDeploymentNameRegexp.MatchString(name) {
	// 		computationService.computations = append(computationService.computations, Computation{Algorithm: Kruskal, Name: name})
	// 	}
	// }

	return computationService, nil
}

func (c *ComputationService) generateComputationName() string {
	return fmt.Sprintf(ComputationDeploymentNamePattern, len(c.computations)+1)
}

func (c *ComputationService) GetAllComputations() []Computation {
	upcxxList := glkube.GetAllDeployments()
	var computations []Computation
	for _, upcxx := range upcxxList.Items {
		computations = append(computations, Computation{
			Name:      upcxx.Spec.StatefulSetName,
			Algorithm: "Prim",
		})
	}

	return computations
}

func (c *ComputationService) GetComputation(name string) Computation {
    upcxx := glkube.GetDeployment(name)
    return Computation{
        Name: upcxx.Spec.StatefulSetName,
        Algorithm: upcxx.Spec.Algorithm,
    }
}

func (c *ComputationService) PostComputation(algorithm glconstants.Algorithm) (Computation, error) {
	computation := Computation{Algorithm: algorithm, Name: c.generateComputationName()}
	if err := glkube.CreateDeployment(computation.Name); err != nil {
		return computation, err
	}

	c.computations = append(c.computations, computation)
	return computation, nil
}
