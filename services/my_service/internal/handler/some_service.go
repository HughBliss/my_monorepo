package handler

import (
	"context"
	"github.com/hughbliss/my_protobuf/gen/someservice"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/hughbliss/my_toolkit/tracer"
	"github.com/pkg/errors"
	globalLog "github.com/rs/zerolog/log"
	"sync"
	"time"
)

func NewSomeServiceHandler() someservice.SomeServiceServer {
	return &SomeServiceHandler{
		report: reporter.InitReporter("SomeServiceHandler"),
	}
}

type SomeServiceHandler struct {
	report reporter.Reporter
}

// SomeExampleMethod implements someservice.SomeServiceServer.
func (s *SomeServiceHandler) SomeExampleMethod(ctx context.Context, _ *someservice.SomeExampleMethodRequest) (*someservice.SomeExampleMethodResponse, error) {
	ctx, log, end := s.report.Start(ctx, "SomeExampleMethod")
	defer end()

	log.Info().Msg("SomeExampleMethod called")
	var wg = &sync.WaitGroup{}

	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := SomeLogic(ctx, 0)
			log.Error().Err(err).Stack().Msg("SomeLogic called")
		}()
	}

	wg.Wait()

	return &someservice.SomeExampleMethodResponse{
		SomeResponse: "ok",
	}, nil
}

func SomeLogic(ctx context.Context, counter uint32) error {

	ctx, span := tracer.Provider.
		Tracer("ExternalLogic").
		Start(ctx, "SomeLogic")
	defer span.End()
	log := globalLog.With().
		Str("service", "ExternalLogic").
		Str("method", "SomeLogic").
		Ctx(ctx).Logger()

	log.Info().Msg("SomeLogic called")
	time.Sleep(10 * time.Millisecond)
	if counter != 100 {
		counter++
		return SomeLogic(ctx, counter)
	}

	log.Warn().Msg("ends with err")
	return errors.New("some error ")
}
