package outfit

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/nats-io/nats.go"

	"github.com/jantytgat/go-kit/slogd"
)

func NewWorkerPool(name string, chMsg chan *nats.Msg, handler MsgHandler, logger *slog.Logger) *WorkerPool {
	w := &WorkerPool{
		name:       name,
		subject:    handler.Subject(),
		chMsg:      chMsg,
		maxWorkers: handler.MaxWorkers(),
		handler:    handler.Handler(),
		logger:     logger,
		workers:    make([]bool, handler.MaxWorkers()),
		wg:         &sync.WaitGroup{},
		mux:        sync.Mutex{},
	}
	return w
}

type WorkerPool struct {
	name       string
	subject    string
	chMsg      chan *nats.Msg
	maxWorkers int
	handler    func(msg *nats.Msg)
	logger     *slog.Logger

	workers             []bool
	workerCtx           context.Context
	workerCtxCancelFunc context.CancelFunc
	wg                  *sync.WaitGroup

	mux sync.Mutex
}

func (w *WorkerPool) Start(ctx context.Context) {
	w.logger.LogAttrs(ctx, slogd.LevelDebug, "starting worker-pool")
	w.workerCtx, w.workerCtxCancelFunc = context.WithCancel(ctx)
	go w.monitor(w.workerCtx, w.handler)
}

func (w *WorkerPool) Shutdown() {
	w.logger.LogAttrs(w.workerCtx, slogd.LevelDebug, "shutting down worker-pool")
	w.workerCtxCancelFunc()
	w.wg.Wait()
}

func (w *WorkerPool) addWorker() (int, error) {
	w.mux.Lock()
	defer w.mux.Unlock()

	for i, _ := range w.workers {
		if !w.workers[i] {
			w.workers[i] = true
			w.wg.Add(1)
			return i, nil
		}
	}
	return 0, errors.New("no free worker slot")
}

func (w *WorkerPool) count() int {
	w.mux.Lock()
	defer w.mux.Unlock()

	running := 0
	for _, active := range w.workers {
		if active {
			running++
		}
	}
	return running
}

func (w *WorkerPool) deleteWorker(id int) {
	defer w.wg.Done()

	w.mux.Lock()
	defer w.mux.Unlock()
	w.workers[id] = false
}

func (w *WorkerPool) id() string {
	return w.name
}

func (w *WorkerPool) monitor(ctx context.Context, f func(msg *nats.Msg)) {
	w.logger.LogAttrs(ctx, slogd.LevelDebug, "starting worker-pool monitor")
	w.wg.Add(1)
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			w.logger.LogAttrs(ctx, slogd.LevelDebug, "stopping worker-pool monitor")
			return
		default:
			if w.count() < w.maxWorkers {
				go w.run(ctx, f)
			}
		}
	}
}

func (w *WorkerPool) run(ctx context.Context, f func(msg *nats.Msg)) {
	var err error
	var id int
	if id, err = w.addWorker(); err != nil {
		return
	}
	workerLogger := w.logger.With(slog.Int("worker-id", id))
	workerLogger.LogAttrs(ctx, slogd.LevelTrace, "starting worker-pool instance")
	defer w.deleteWorker(id)

forLoop:
	for {
		select {
		case <-ctx.Done():
			workerLogger.LogAttrs(ctx, slogd.LevelTrace, "draining worker-pool", slog.Int("queue-length", len(w.chMsg)))
			if len(w.chMsg) == 0 {
				break forLoop
			}
		case msg := <-w.chMsg:
			workerLogger.LogAttrs(ctx, slogd.LevelTrace, "worker received message")
			f(msg)
		}
	}
	workerLogger.LogAttrs(ctx, slogd.LevelTrace, "stopping worker-pool instance")
}

func (w *WorkerPool) updateLogger(logger *slog.Logger) *slog.Logger {
	w.logger = logger
	return w.logger
}
