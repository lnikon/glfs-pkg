package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"

	glconstants "github.com/lnikon/glfs-pkg/pkg/constants"
)

// /algorithm endpoint
type AlgorithmRequest struct {
}

type algorithmResponse struct {
	Algorithm []glconstants.Algorithm
}

func MakeAlgorithmEndpoint(svc *AlgorithmService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		a := svc.Algorithm()
		return algorithmResponse{a}, nil
	}
}

func DecodeAlgorithmRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return AlgorithmRequest{}, nil
}

type GetComputationRequest struct {
	Name string
}

type GetComputationResponse struct {
	Computation *Computation
}

func MakeGetComputationEndpoint(svc ComputationServiceIfc) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(GetComputationRequest)
		computation, err := svc.GetComputation(req.Name)
		if err != nil {
			return nil, err
		}

		return GetComputationResponse{Computation: computation}, nil
	}
}

func DecodeGetComputationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	name := mux.Vars(r)["name"]
	return GetComputationRequest{
		Name: name,
	}, nil
}

type GetAllComputationsRequest struct {
}

type GetAllComputationsResponse struct {
	Computations []Computation
}

func MakeGetAllComputationsEndpoint(svc ComputationServiceIfc) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		return GetAllComputationsResponse{Computations: svc.GetAllComputations()}, nil
	}
}

func DecodeGetAllComputationsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return GetAllComputationsRequest{}, nil
}

// Universal encoder for all responses
func EncodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

type PostComputationRequest struct {
	Algorithm glconstants.Algorithm
}

type PostComputationResponse struct {
	Computation *Computation
}

func MakePostComputationEndpoint(svc ComputationServiceIfc) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(PostComputationRequest)
		computation, err := svc.PostComputation(req.Algorithm)
		if err != nil {
			return nil, err
		}

		return PostComputationResponse{Computation: computation}, nil
	}
}

func DecodePostComputationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body struct {
		Algorithm glconstants.Algorithm `json:"algorithm"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return PostComputationRequest{
		Algorithm: body.Algorithm,
	}, nil
}
