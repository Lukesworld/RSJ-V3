package workerpool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// Job represents a unit of work to be processed.
type Job interface {
	Process(ctx context.Context) error
	ID() string
}

// Result captures the outcome of a job.
type Result struct {
	JobID     string
	WorkerID  int
	Error     error
	Duration  time.Duration
	Timestamp time.Time
}

// Pool manages a pool of workers.
type Pool struct {
	numWorkers int
	jobQueue   chan Job
	results    chan Result
	wg         sync.WaitGroup
	quit       chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc

	// Metrics
	processedCount int64
	errorCount     int64
}

// NewPool creates a new worker pool.
func NewPool(numWorkers int, bufferSize int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		numWorkers: numWorkers,
		jobQueue:   make(chan Job, bufferSize),
		results:    make(chan Result, bufferSize),
		quit:       make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start initializes the workers and starts processing jobs.
func (p *Pool) Start() {
	log.Printf("[WorkerPool] Starting %d workers...", p.numWorkers)
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	// Start a goroutine to handle results/metrics logging if needed
	go p.monitorResults()
}

// Stop gracefully shuts down the pool, waiting for workers to finish.
func (p *Pool) Stop() {
	log.Println("[WorkerPool] Stopping pool...")
	p.cancel() // Signal context cancellation
	close(p.jobQueue) // Close job queue to stop accepting new jobs
	p.wg.Wait() // Wait for all workers to finish
	close(p.results)
	log.Println("[WorkerPool] Pool stopped.")
}

// Submit adds a job to the queue. Returns error if pool is stopped or full (if non-blocking add implemented).
func (p *Pool) Submit(job Job) {
	select {
	case p.jobQueue <- job:
		// Job submitted
	case <-p.ctx.Done():
		log.Printf("[WorkerPool] Cannot submit job %s: pool stopping", job.ID())
	}
}

// worker is the main loop for each worker goroutine.
func (p *Pool) worker(id int) {
	defer p.wg.Done()
	log.Printf("[WorkerPool] Worker %d started", id)

	for job := range p.jobQueue {
		// Check context before processing
		if p.ctx.Err() != nil {
			log.Printf("[WorkerPool] Worker %d stopping (context cancelled)", id)
			return
		}

		startTime := time.Now()
		// Safe execution with panic recovery
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic in job %s: %v", job.ID(), r)
					log.Printf("[WorkerPool] CRITICAL: Worker %d panicked: %v", id, r)
				}
			}()
			err = job.Process(p.ctx)
		}()

		duration := time.Since(startTime)
		atomic.AddInt64(&p.processedCount, 1)
		if err != nil {
			atomic.AddInt64(&p.errorCount, 1)
			log.Printf("[WorkerPool] Job %s failed: %v", job.ID(), err)
		}

		// Send result
		select {
		case p.results <- Result{
			JobID:     job.ID(),
			WorkerID:  id,
			Error:     err,
			Duration:  duration,
			Timestamp: time.Now(),
		}:
		default:
			// Result channel full, maybe log warning?
		}
	}
	log.Printf("[WorkerPool] Worker %d stopped", id)
}

func (p *Pool) monitorResults() {
	for res := range p.results {
		if res.Error != nil {
			// Could implement advanced error handling/retry logic here
		}
	}
}

// Stats returns current metrics.
func (p *Pool) Stats() (processed int64, errors int64) {
	return atomic.LoadInt64(&p.processedCount), atomic.LoadInt64(&p.errorCount)
}
