package pool

import (
	spxLog "gitbuh.com/spxzx/spxgo/log"
	"time"
)

type Worker struct {
	pool     *Pool
	task     chan func() // 任务队列
	lastTime time.Time   // 执行任务的最后时间
}

func (w *Worker) run() {
	w.pool.incRunning()
	go w.running()
}

func (w *Worker) running() {
	// 错误捕获
	defer func() {
		w.pool.decRunning()
		w.pool.workerCache.Put(w)
		if err := recover(); err != nil {
			if w.pool.PanicHandler != nil {
				w.pool.PanicHandler()
			} else {
				spxLog.Default().Error(err)
			}
		}
		w.pool.cond.Signal()
	}()
	for t := range w.task {
		if t == nil {
			return // 只要不被 return 就说明该 worker 一直处在工作状态中，所以不能进行 decRunning
		}
		t()
		// 任务完成 worker 空闲返还
		w.pool.PutWorker(w)
		// w.pool.decRunning()
	}
}
