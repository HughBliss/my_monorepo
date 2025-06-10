package handler

import (
	"context"
	someservicev1 "github.com/hughbliss/my_protobuf/go/pkg/gen/someservice/v1"
	"github.com/hughbliss/my_toolkit/reporter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

func NewSomeServiceHandler(logic SomeLogicProvider) someservicev1.SomeServiceServer {
	return &SomeServiceHandler{
		logic:  logic,
		report: reporter.InitReporter("SomeServiceHandler"),
	}
}

type SomeServiceHandler struct {
	report reporter.Reporter
	logic  SomeLogicProvider
}

func (s *SomeServiceHandler) AnotherExampleMethod(ctx context.Context, request *someservicev1.SomeExampleMethodRequest) (*someservicev1.SomeExampleMethodResponse, error) {
	//TODO implement me
	panic("implement me")
}

type SomeLogicProvider interface {
	SomeLogic(ctx context.Context, counter uint32) error
}

// SomeExampleMethod implements someservice.SomeServiceServer.
func (s *SomeServiceHandler) SomeExampleMethod(ctx context.Context, req *someservicev1.SomeExampleMethodRequest) (*someservicev1.SomeExampleMethodResponse, error) {
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

	return nil, status.Errorf(codes.ResourceExhausted, "resource exhausted")
}
