package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Scheduler struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{ctx: ctx, cancel: cancel}
}

func (s *Scheduler) StartDaily(name string, scheduledAt string, job func(context.Context) error) error {
	hour, minute, err := parseHourMinute(scheduledAt)
	if err != nil {
		return err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for {
			now := time.Now()
			next := nextDailyRun(now, hour, minute)
			timer := time.NewTimer(time.Until(next))

			select {
			case <-s.ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				log.Printf("running scheduled job %s", name)
				if err := job(s.ctx); err != nil {
					log.Printf("scheduled job %s failed: %v", name, err)
				}
			}
		}
	}()

	log.Printf("scheduled job %s daily at %02d:%02d", name, hour, minute)
	return nil
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
}

func parseHourMinute(value string) (int, int, error) {
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid daily schedule time %q: %w", value, err)
	}

	return parsed.Hour(), parsed.Minute(), nil
}

func nextDailyRun(now time.Time, hour int, minute int) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}

	return next
}
