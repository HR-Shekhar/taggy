package aigen

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	// ErrQueueFull is returned when the bounded job queue cannot accept more work.
	ErrQueueFull = errors.New("ai generation queue is full")
	// ErrPoolClosed is returned after Shutdown.
	ErrPoolClosed = errors.New("ai generation pool is closed")
)

// Job is a unit of AI work. The context is cancelled on pool shutdown or job timeout.
type Job func(ctx context.Context)

// Config tunes the bounded worker pool.
type Config struct {
	// Workers caps concurrent OpenRouter HTTP clients (default 2).
	Workers int
	// QueueSize is the buffered backlog before Submit returns ErrQueueFull (default 64).
	QueueSize int
	// JobTimeout bounds a single generation including provider retries (default 15m).
	JobTimeout time.Duration
}

// Pool runs AI jobs with a fixed worker count so request handlers never spawn unbounded goroutines.
type Pool struct {
	jobs       chan Job
	workers    int
	jobTimeout time.Duration
	log        zerolog.Logger

	rootCtx    context.Context
	cancelRoot context.CancelFunc
	wg         sync.WaitGroup

	mu     sync.Mutex
	closed bool
}

// NewPool creates an idle pool. Call Start before Submit.
func NewPool(cfg Config, log zerolog.Logger) *Pool {
	workers := cfg.Workers
	if workers <= 0 {
		workers = 2
	}
	queue := cfg.QueueSize
	if queue <= 0 {
		queue = 64
	}
	timeout := cfg.JobTimeout
	if timeout <= 0 {
		timeout = 15 * time.Minute
	}
	rootCtx, cancel := context.WithCancel(context.Background())
	return &Pool{
		jobs:       make(chan Job, queue),
		workers:    workers,
		jobTimeout: timeout,
		log:        log,
		rootCtx:    rootCtx,
		cancelRoot: cancel,
	}
}

// Start launches worker goroutines. Safe to call once.
func (p *Pool) Start() {
	p.log.Info().
		Int("workers", p.workers).
		Int("queue", cap(p.jobs)).
		Dur("job_timeout", p.jobTimeout).
		Msg("ai generation pool started")
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()
	for {
		select {
		case <-p.rootCtx.Done():
			return
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			p.runJob(id, job)
		}
	}
}

func (p *Pool) runJob(workerID int, job Job) {
	ctx, cancel := context.WithTimeout(p.rootCtx, p.jobTimeout)
	defer cancel()
	started := time.Now()
	defer func() {
		if rec := recover(); rec != nil {
			p.log.Error().
				Interface("panic", rec).
				Int("worker", workerID).
				Msg("ai generation job panicked")
		}
	}()
	job(ctx)
	p.log.Debug().
		Int("worker", workerID).
		Dur("elapsed", time.Since(started)).
		Msg("ai generation job finished")
}

// Submit enqueues a job. Returns ErrQueueFull if the backlog is saturated.
func (p *Pool) Submit(job Job) error {
	if job == nil {
		return nil
	}
	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()
	if closed {
		return ErrPoolClosed
	}
	select {
	case p.jobs <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

// Shutdown cancels in-flight jobs and waits for workers to exit (or ctx).
func (p *Pool) Shutdown(ctx context.Context) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()

	p.cancelRoot()
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		p.log.Info().Msg("ai generation pool stopped")
	case <-ctx.Done():
		p.log.Warn().Msg("ai generation pool shutdown timed out; workers may still be exiting")
	}
}
