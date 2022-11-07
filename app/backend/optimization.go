package backend

import (
	"runtime"
	"time"
)

type IOptimizationService interface {
	Start()
}

type OptimizationServiceImpl struct {
}

func (c *OptimizationServiceImpl) Start() {
	go func() {
		for {
			time.Sleep(2 * time.Minute)
			runtime.GC()
		}
	}()
}
