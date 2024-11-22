package server

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type WorkerPool struct {
	tasks   chan func(ctx context.Context) error // 处理的任务方法
	wg      sync.WaitGroup                       // 处理子协程
	ctx     context.Context                      // 全文的上下文
	cancel  context.CancelFunc                   // 取消函数
	stopped bool                                 // 标识worker pool是否已经停止
	log     *zap.SugaredLogger
}

// 初始化worker pool
func NewWorkerPool(ctx context.Context,
	log *zap.SugaredLogger) (*WorkerPool, error) {
	w := &WorkerPool{
		tasks: make(chan func(ctx context.Context) error, 256),
		log:   log,
	}

	w.ctx, w.cancel = context.WithCancel(ctx)

	return w, nil
}

// 开始
func (wp *WorkerPool) Start() {
	go func() {
		for task := range wp.tasks {
			if task == nil {
				wp.log.Info("worker pool has been close ...")
				return
			}

			wp.wg.Add(1)
			go func(f func(ctx context.Context) error) {
				defer wp.wg.Done()
				_ = f(wp.ctx)
			}(task)
		}
	}()
}

// 停止
func (wp *WorkerPool) Stop() {
	wp.cancel()
	wp.wg.Wait()
	close(wp.tasks)
	wp.stopped = true
}

// 提交任务至任务池
func (wp *WorkerPool) Submit(task func(ctx context.Context) error) error {
	if wp.stopped {
		return fmt.Errorf("worker pool has been stopped")
	}
	select {
	case wp.tasks <- task:
		return nil
	case <-wp.ctx.Done():
		return wp.ctx.Err()
	}
}
