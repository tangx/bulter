package butler

import (
	"context"
	"log"
	"os"
	"runtime"
)

type Butler struct {
	workers     int
	workerQueue chan *worker
	jobs        int
	jobQueue    chan func()

	ctx context.Context
}

// Default return a butler , which workers' number is equal GOMAXPROC
// and jobs' number is double of workers
// https://golang.org/doc/effective_go#parallel
func Default() *Butler {
	b := &Butler{}
	b.initial()
	return b
}

// Init return a new Butler
func (b *Butler) Init(funcs ...OptionFunc) {
	for _, fn := range funcs {
		fn(b)
	}

	b.initial()
}

// Work start
func (b *Butler) Work() {
	// https://colobu.com/2015/10/09/Linux-Signals/
	sigs := make(chan os.Signal, 1)
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

Loop:
	for {
		select {
		case <-sigs:
			break Loop
		case worker := <-b.workerQueue:
			job := <-b.jobQueue
			go b.assign(worker, job)
		}
	}

}

func (b *Butler) AddJobs(funcs ...func()) {
	for _, fn := range funcs {
		b.jobQueue <- fn
	}
}

// SetDefaults set default value for butler
func (b *Butler) SetDefaults() {
	b.workers = runtime.GOMAXPROCS(0)
	b.jobs = b.workers * 2
}

// initial butler
func (b *Butler) initial() {
	b.SetDefaults()

	b.jobQueue = make(chan func(), b.jobs)

	b.workerQueue = make(chan *worker, b.workers)
	for i := 0; i < b.workers; i++ {
		log.Println("register a new worker")
		b.workerQueue <- newWorker()
	}

}

// assign a job to a worker
func (b *Butler) assign(w *worker, job func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}

		b.workerQueue <- w
		log.Println("<<< job done, re-assgined")
	}()

	w.do(job)
}

// OptionFunc
type OptionFunc = func(b *Butler)

// WithWorkers set concurrency worker numbers
func WithWorkers(n int) OptionFunc {
	return func(b *Butler) {
		b.workers = n
	}
}

// WithJobs set max jobs queue
func WithJobs(n int) OptionFunc {
	return func(b *Butler) {
		b.jobs = n
	}
}
