package tasks

import (
	"errors"
	"sync"
	"time"
)

const (
	defaultTimeout = 1 * time.Second
)

var (
	ErrAlreadyStopped = errors.New("task already has stopped")
)

type ExeOption func(exe *PeriodicExecutor)

type PeriodicExecutor struct {
	stopChan       chan bool
	lock           sync.RWMutex
	goroutines     int
	taskRunning    bool
	timeout        time.Duration
	initialTimeout time.Duration
	f              func() bool
}

func (pe *PeriodicExecutor) SetSchedulerTimeout(duration time.Duration) *PeriodicExecutor {
	pe.timeout = duration
	return pe
}

func (pe *PeriodicExecutor) StartTask(function func() bool) *PeriodicExecutor {
	pe.lock.Lock()
	defer pe.lock.Unlock()

	pe.f = function

	if !pe.taskRunning {
		pe.taskRunning = true
		pe.startSchedulerTask()
	}

	return pe
}

func (pe *PeriodicExecutor) StopTask() {
	pe.lock.Lock() //await
	defer pe.lock.Unlock()

	if !pe.taskRunning {
		panic(ErrAlreadyStopped)
	}

	pe.taskRunning = false
	close(pe.stopChan)
}

func (pe *PeriodicExecutor) startSchedulerTask() *PeriodicExecutor {
	for i := 0; i < pe.goroutines; i++ {
		go func() {
			if pe.initialTimeout > 0 {
				if continueExec := pe.wait(pe.initialTimeout); !continueExec {
					return
				}
			}
			for {
				pe.lock.RLock()

				if !pe.taskRunning {
					pe.lock.RUnlock()
					break
				}

				continueExec := pe.f()

				pe.lock.RUnlock()

				if !continueExec {
					if continueExec = pe.wait(pe.timeout); continueExec {
						continue
					} else {
						break
					}
				}
			}
		}()
	}
	return pe
}

func (pe *PeriodicExecutor) wait(t time.Duration) bool {
	select {
	case <-time.After(t):
		return true
	case <-pe.stopChan:
		return false
	}
}

func NewPeriodicExecutor(timeout time.Duration, opts ...ExeOption) *PeriodicExecutor {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	exe := &PeriodicExecutor{
		timeout:    timeout,
		stopChan:   make(chan bool),
		goroutines: 1,
	}
	for _, o := range opts {
		o(exe)
	}
	return exe
}

func ExecutorWithInitialTimeout(initialTimeout time.Duration) ExeOption {
	return func(exe *PeriodicExecutor) {
		exe.initialTimeout = initialTimeout
	}
}

func ExecutorWithGoroutinesCount(count int) ExeOption {
	return func(exe *PeriodicExecutor) {
		exe.goroutines = count
	}
}
