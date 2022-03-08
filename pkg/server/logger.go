package server

import (
	"fmt"
	"time"

	log "github.com/go-kit/log"
	glconstants "github.com/lnikon/glfs-pkg/pkg/constants"
)

type LoggingMiddleware struct {
	Next   ComputationServiceIfc
	Logger log.Logger
}

func (mw LoggingMiddleware) GetComputation(name string) (computation *Computation, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "GetComputation",
			"input", fmt.Sprintf("%v", name),
			"output", fmt.Sprintf("%v", computation),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	computation, err = mw.Next.GetComputation(name)
	if err != nil {
		mw.Logger.Log("Error: ", err.Error())
	}
	return
}

func (mw LoggingMiddleware) GetAllComputations() (output []Computation) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "GetAllComputations",
			"output", fmt.Sprintf("%v", output),
			"took", time.Since(begin),
		)
	}(time.Now())

	output = mw.Next.GetAllComputations()
	return
}

func (mw LoggingMiddleware) PostComputation(algorithm glconstants.Algorithm) (output *Computation, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "PostComputation",
			"input", fmt.Sprintf("%v", algorithm),
			"output", fmt.Sprintf("%v", output),
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.PostComputation(algorithm)
	if err != nil {
		mw.Logger.Log("Error: ", err.Error())
	}
	return
}

func (mw LoggingMiddleware) DeleteComputation(name string) (err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "DeleteComputation",
			"input", fmt.Sprintf("%v", name),
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.DeleteComputation(name)
	if err != nil {
		mw.Logger.Log("Error: ", err.Error())
	}
	return
}
