package bulter

import "log"

type Bulter struct {
	workers     int
	workerQueue chan *worker
	jobs        int
	jobQueue    chan func()
}

// New return a new Bulter
func New(funcs ...OptionFunc) *Bulter {
	b := &Bulter{}

	for _, fn := range funcs {
		fn(b)
	}

	b.initial()
	return b
}

// Work start
func (b *Bulter) Work() {

	for worker := range b.workerQueue {
		job := <-b.jobQueue
		log.Println(">>> a worker accpet a new job")
		go b.assign(worker, job)
	}
}

func (b *Bulter) AddJobs(funcs ...func()) {
	for _, fn := range funcs {
		b.jobQueue <- fn
	}
}

// initial bulter
func (b *Bulter) initial() {
	if b.jobs == 0 {
		b.jobs = 20
	}
	b.jobQueue = make(chan func(), b.jobs)

	if b.workers == 0 {
		b.workers = 5
	}
	b.workerQueue = make(chan *worker, b.workers)
	for i := 0; i < b.workers; i++ {
		log.Println("register a new worker")
		b.workerQueue <- newWorker()
	}
}

// assign a job to a worker
func (b *Bulter) assign(w *worker, job func()) {
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
type OptionFunc = func(b *Bulter)

// WithWorkers set concurrency worker numbers
func WithWorkers(n int) OptionFunc {
	return func(b *Bulter) {
		b.workers = n
	}
}

// WithJobs set max jobs queue
func WithJobs(n int) OptionFunc {
	return func(b *Bulter) {
		b.jobs = n
	}
}
