package usecase

import (
	"context"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/pkg/errors"
	"math/rand"
	"time"
)

type SomeUsecase struct {
	report reporter.Reporter
}

func NewSomeUsecase() *SomeUsecase {
	return &SomeUsecase{
		report: reporter.InitReporter("SomeUsecase"),
	}
}

func (s SomeUsecase) SomeLogic(ctx context.Context, counter uint32) error {
	ctx, log, end := s.report.Start(ctx, "SomeLogic")
	defer end()

	log.Info().Msg("SomeLogic called")
	time.Sleep(10 * time.Millisecond)
	if counter != 5 {
		counter++
		if err := s.SomeLogic(ctx, counter); err != nil {
			log.Error().Err(err).Msgf("SomeLogic failed %d", counter)
			return err
		}
	}

	if rand.Int()%2 == 0 {
		return errors.New("random error")
	}

	return nil
}
