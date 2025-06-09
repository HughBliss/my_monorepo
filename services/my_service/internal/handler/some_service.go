package handler

import (
	"context"
	"github.com/hughbliss/my_protobuf/gen/someservice"
	"github.com/hughbliss/my_toolkit/reporter"
	"sync"
)

func NewSomeServiceHandler(logic SomeLogicProvider) someservice.SomeServiceServer {
	return &SomeServiceHandler{
		logic:  logic,
		report: reporter.InitReporter("SomeServiceHandler"),
	}
}

type SomeServiceHandler struct {
	report reporter.Reporter
	logic  SomeLogicProvider
}

type SomeLogicProvider interface {
	SomeLogic(ctx context.Context, counter uint32) error
}

// SomeExampleMethod implements someservice.SomeServiceServer.
func (s *SomeServiceHandler) SomeExampleMethod(ctx context.Context, req *someservice.SomeExampleMethodRequest) (*someservice.SomeExampleMethodResponse, error) {
	ctx, log, end := s.report.Start(ctx, "SomeExampleMethod")
	defer end()

	log.Info().Any("request", req).Send()
	var wg = &sync.WaitGroup{}

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.logic.SomeLogic(ctx, 0); err != nil {
				log.Error().Stack().Err(err).Msg("error while running SomeLogic")
				return
			}
			log.Info().Msg("SomeLogic finished with success")
		}()
	}

	wg.Wait()

	return &someservice.SomeExampleMethodResponse{
		SomeResponse: "ok",
	}, nil
}
