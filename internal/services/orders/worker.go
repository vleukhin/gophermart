package orders

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/vleukhin/gophermart/internal/services/accrual"
	"time"
)

type (
	Worker struct {
		accrual accrual.Service
		in      chan job
		out     chan accrual.OrderInfo
	}
	job struct {
		OrderID  string
		Try      int
		MaxTries int
		Delay    time.Duration
	}
)

func newWorker(accrual accrual.Service, in chan job, out chan accrual.OrderInfo) *Worker {
	return &Worker{
		accrual: accrual,
		in:      in,
		out:     out,
	}
}

func (w *Worker) Run() {
	for job := range w.in {
		log.Info().Str("order", job.OrderID).Msg("Getting order info from accrual service")
		info, err := w.accrual.GetOrderInfo(job.OrderID)
		if err != nil {
			log.Error().Err(err).Str("order", job.OrderID).Msg("Failed to get order info")
			if err := job.retry(); err != nil {
				log.Error().Err(err).Str("order", job.OrderID).Msg("")
				continue
			}
			time.AfterFunc(job.Delay, func() {
				w.in <- job
			})
			continue
		}

		w.out <- info
	}
	log.Info().Msg("Order worker stopped")
}

func (j *job) retry() error {
	if j.Try == 0 {
		j.Delay = 300 * time.Millisecond
	} else {
		j.Delay *= 2
	}
	j.Try++
	if j.Try >= j.MaxTries {
		return fmt.Errorf("order info job failed after %d tries", j.Try)
	}

	return nil
}

func newJob(orderID string, retries int) job {
	return job{
		OrderID:  orderID,
		MaxTries: retries,
	}
}
