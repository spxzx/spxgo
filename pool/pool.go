package pool

import (
	"errors"
	"gitbuh.com/spxzx/spxgo/config"
	"sync"
	"sync/atomic"
	"time"
)

const DefaultExpire = 1

var (
	ErrorInValidCap    = errors.New("cap can not < 0")
	ErrorInValidExpire = errors.New("expire time can not < 0")
	ErrorPoolClosed    = errors.New("pool has been released")
	ErrorRestart       = errors.New("pool restart error")
)

type sign struct {
}

type Pool struct {
	workers      []*Worker     // 空闲worker
	workerCache  sync.Pool     // worker 缓存
	cap          int32         // pool 最大容量
	running      int32         // 正在运行 worker 数量
	expire       time.Duration // 过期时间 空闲 worker 超过该时间就回收
	release      chan sign     // 释放资源 pool 就不能使用了
	lock         sync.Mutex    // 保护 pool 中相关资源的安全
	once         sync.Once     // 保证 release 只能调用一次，不能多次调用
	cond         *sync.Cond    // 基于互斥锁/读写锁实现的条件变量,协调想要访问共享资源的那些 Goroutine
	PanicHandler func()
}

func NewPool(cap int) (*Pool, error) {
	return NewTimePool(cap, DefaultExpire)
}

func NewTimePool(cap int, expire int) (*Pool, error) {
	if cap <= 0 {
		return nil, ErrorInValidCap
	}
	if expire <= 0 {
		return nil, ErrorInValidExpire
	}
	p := &Pool{
		cap:     int32(cap),
		expire:  time.Duration(expire) * time.Second,
		release: make(chan sign, 1),
	}
	p.workerCache.New = func() any {
		return &Worker{
			pool: p,
			task: make(chan func()),
		}
	}
	p.cond = sync.NewCond(&p.lock)
	go p.expireWorker()
	return p, nil
}

func NewPoolWithConfig() (*Pool, error) {
	capacity, ok1 := config.Conf.Pool["cap"].(int64)
	if !ok1 {
		return nil, errors.New("cap config not exit")
	}
	expire, ok2 := config.Conf.Pool["expire"].(int64)
	if !ok2 {
		return nil, errors.New("expire config not exit")
	}
	return NewTimePool(int(capacity), int(expire))
}

// 防止大量 worker 的占用，造成浪费性能
func (p *Pool) expireWorker() {
	// 定时清理过期的 free worker
	ticker := time.NewTicker(p.expire)
	for range ticker.C {
		if p.IsReleased() {
			break
		}
		// 循环空闲 worker 如果当前时间和worker最后运行任务的时间 差值大于 expire 就进行清理
		p.lock.Lock()
		idleWorkers := p.workers
		n := len(idleWorkers) - 1
		if n >= 0 {
			clearN := -1
			for i, w := range idleWorkers {
				if time.Now().Sub(w.lastTime) <= p.expire {
					break
				}
				clearN = i
				w.task <- nil
				idleWorkers[i] = nil
			}
			if clearN != -1 {
				if clearN >= len(idleWorkers)-1 {
					p.workers = idleWorkers[:0]
				} else {
					p.workers = idleWorkers[clearN+1:]
				}
			}
			// fmt.Printf("清除完成,running: %d, workers: %v \n", p.running, p.workers)
		}
		p.lock.Unlock()
	}
}

// 代码复用 - 没有空闲 worker 新建一个 worker
func (p *Pool) createWorker() *Worker {
	wc := p.workerCache.Get()
	var w *Worker
	if wc == nil {
		w = &Worker{
			pool: p,
			task: make(chan func(), 4),
		}
	} else {
		w = wc.(*Worker)
	}
	w.run()
	return w
}

func (p *Pool) GetWorker() *Worker {
	// 1. 获取 pool 里面的 worker
	// 2. 有空闲的 worker 直接获取
	p.lock.Lock() // 注意锁的位置
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n >= 0 {
		w := idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
		return w
	}
	// 3. 没有空闲 worker 新建一个 worker
	if p.running < p.cap {
		p.lock.Unlock()
		return p.createWorker()
	}
	// 4. 如果正在运行的 worker > pool.cap ，阻塞等待 worker 释放
	//for {
	p.lock.Unlock()
	return p.waitIdleWorker()
	//}
}

func (p *Pool) waitIdleWorker() *Worker {
	p.lock.Lock()
	p.cond.Wait()
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n < 0 {
		p.lock.Unlock()
		// 防止出现panic等错误时造成程序一直阻塞
		if p.running < p.cap {
			return p.createWorker()
		}
		return p.waitIdleWorker()
	}
	w := idleWorkers[0]
	idleWorkers[n] = nil
	p.workers = idleWorkers[:n]
	p.lock.Unlock()
	return w
}

// Submit 提交任务
func (p *Pool) Submit(task func()) error {
	if len(p.release) > 0 {
		return ErrorPoolClosed
	}
	// 获取Pool里面一个worker，然后执行任务
	w := p.GetWorker()
	w.task <- task
	return nil
}

func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *Pool) PutWorker(w *Worker) {
	w.lastTime = time.Now()
	p.lock.Lock()
	p.workers = append(p.workers, w)
	p.cond.Signal()
	p.lock.Unlock()
}

func (p *Pool) Release() {
	p.once.Do(func() {
		// 只执行一次
		p.lock.Lock()
		workers := p.workers
		for i, w := range workers {
			w.task = nil
			w.pool = nil
			workers[i] = nil
		}
		p.workers = nil
		p.lock.Unlock()
		p.release <- sign{}
	})
}

func (p *Pool) IsReleased() bool {
	return len(p.release) > 0
}

func (p *Pool) Restart() error {
	if p.IsReleased() {
		return ErrorRestart
	}
	_ = <-p.release
	go p.expireWorker()
	return nil
}

func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *Pool) Free() int {
	return int(p.cap - p.running)
}
