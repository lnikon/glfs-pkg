package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
)

type algorithmRequest struct {
}

type algorithmResponse struct {
	Algorithm []Algorithm
}

func MakeAlgorithmEndpoint(svc *AlgorithmService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		a := svc.Algorithm()
		return algorithmResponse{a}, nil
	}
}

func DecodeAlgorithmRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return algorithmRequest{}, nil
}

type GetAllComputationsRequest struct {
}

type GetAllComputationsResponse struct {
	Computations []Computation
}

func MakeGetAllComputationsEndpoint(svc *ComputationService) endpoint.Endpoint {
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
	Algorithm Algorithm
}

type PostComputationResponse struct {
	Computation Computation
}

func MakePostComputationEndpoint(svc *ComputationService) endpoint.Endpoint {
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
		Algorithm Algorithm `json:"algorithm"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return PostComputationRequest{
		Algorithm: body.Algorithm,
	}, nil
}
