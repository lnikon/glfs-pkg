package server

import (
	"fmt"

	glconstants "github.com/lnikon/glfs-pkg/pkg/constants"
	glkube "github.com/lnikon/glfs-pkg/pkg/kube"
)

const (
	ComputationDeploymentNamePattern = "computation-%d"
)

type Computation struct {
	Algorithm glconstants.Algorithm `json:"algorithm"`
	Name      string                `json:"name"`
}

func (c *Computation) String() string {
	return fmt.Sprintf("{Algorithm: %v, Name: %v}", c.Algorithm, c.Name)
}

type ComputationServiceIfc interface {
	GetComputation(name string) (*Computation, error)
	GetAllComputations() []Computation
	PostComputation(algorithm glconstants.Algorithm) (*Computation, error)
}

type ComputationService struct {
	computations []Computation
}

func NewComputationService() (ComputationServiceIfc, error) {
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

func (c *ComputationService) GetComputation(name string) (*Computation, error) {
	upcxx := glkube.GetDeployment(name)
	if upcxx == nil {
		return nil, fmt.Errorf("resource does not exists")
	}

	return &Computation{
		Name:      upcxx.Spec.StatefulSetName,
		Algorithm: upcxx.Spec.Algorithm,
	}, nil
}

func (c *ComputationService) PostComputation(algorithm glconstants.Algorithm) (*Computation, error) {
	computation := Computation{Algorithm: algorithm, Name: c.generateComputationName()}
	if err := glkube.CreateDeployment(computation.Name); err != nil {
		return &computation, err
	}

	c.computations = append(c.computations, computation)
	return &computation, nil
}
