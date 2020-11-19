package core

import (
	"context"
	"log"
	"sync"
	"time"
)

type Task func()

// boss 老板
type GoPool struct {
	MaxWorkerIdleTime time.Duration // worker 最大空闲时间
	MaxWorkerNum      int32         // 协程最大数量
	TaskEntryChan     chan *Task    // 任务入列
	Workers           []*worker     // 已创建worker
	FreeWorkerChan    chan *worker  // 空闲worker
	Lock              sync.Mutex
}

const (
	WorkerStatusStop = 1
	WorkerStatusLive = 0
)

// 干活的人
type worker struct {
	Pool         *GoPool
	StartTime    time.Time  // 开始时间
	TaskChan     chan *Task // 执行队列
	LastWorkTime time.Time  // 最后执行时间
	Ctx          context.Context
	Cancel       context.CancelFunc
	Status       int32 // 被过期删掉的标记
}

var defaultPool = func() *GoPool {
	return NewPool()
}()

// 初始化
func NewPool() *GoPool {
	g := &GoPool{
		MaxWorkerIdleTime: 10 * time.Second,
		MaxWorkerNum:      20,
		TaskEntryChan:     make(chan *Task, 2000),
		FreeWorkerChan:    make(chan *worker, 2000),
	}

	// 分发任务
	go g.dispatchTask()

	//清理空闲worker
	go g.fireWorker()

	return g
}

// 定期清理空闲worker
func (g *GoPool) fireWorker() {
	for {
		select {
		// 10秒执行一次
		case <-time.After(10 * time.Second):
			for k, w := range g.Workers {
				if time.Now().Sub(w.LastWorkTime) > g.MaxWorkerIdleTime {
					log.Printf("overtime %v %p", k, w)
					// 终止协程
					w.Cancel()
					// 清理Free
					w.Status = WorkerStatusStop
				}
			}

			g.Lock.Lock()
			g.Workers = g.cleanWorker(g.Workers)
			g.Lock.Unlock()
		}
	}
}

// 递归清理无用worker
func (g *GoPool) cleanWorker(workers []*worker) []*worker {
	for k, w := range workers {
		if time.Now().Sub(w.LastWorkTime) > g.MaxWorkerIdleTime {
			workers = append(workers[:k], workers[k+1:]...) // 删除中间1个元素
			return g.cleanWorker(workers)
		}
	}

	return workers
}

// 分发任务
func (g *GoPool) dispatchTask() {

	for {
		select {
		case t := <-g.TaskEntryChan:
			log.Printf("dispatch task %p", t)
			// 获取worker
			w := g.fetchWorker()
			// 将任务扔给worker
			w.accept(t)
		}
	}
}

// 获取可用worker
func (g *GoPool) fetchWorker() *worker {
	for {
		select {
		// 获取空闲worker
		case w := <-g.FreeWorkerChan:
			if w.Status == WorkerStatusLive {
				return w
			}
		default:
			// 创建新的worker
			if int32(len(g.Workers)) < g.MaxWorkerNum {
				w := &worker{
					Pool:         g,
					StartTime:    time.Now(),
					LastWorkTime: time.Now(),
					TaskChan:     make(chan *Task, 1),
					Ctx:          context.Background(),
					Status:       WorkerStatusLive,
				}
				ctx, cancel := context.WithCancel(w.Ctx)

				w.Cancel = cancel
				// 接到任务自己去执行吧
				go w.execute(ctx)

				g.Lock.Lock()
				g.Workers = append(g.Workers, w)
				g.Lock.Unlock()

				g.FreeWorkerChan <- w

				log.Printf("worker create %p", w)
			}
		}
	}
}

// 添加任务
func (g *GoPool) addTask(t Task) {
	// 将任务放到入口任务队列
	g.TaskEntryChan <- &t
}

// 接受任务
func (w *worker) accept(t *Task) {
	// 每个worker自己的工作队列
	w.TaskChan <- t
}

// 执行任务
func (w *worker) execute(ctx context.Context) {
	for {
		select {
		case t := <-w.TaskChan:
			// 执行
			(*t)()
			// 记录工作状态
			w.LastWorkTime = time.Now()
			w.Pool.FreeWorkerChan <- w
		case <-ctx.Done():
			log.Printf("worker done %p", w)
			return
		}
	}
}

// 执行
func SafeGo(t Task) {
	defaultPool.addTask(t)
}
