package tests

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core/logger"
	"runtime"
	"time"
)

type ITestStatsService interface {
	Start(testName string)
}

type TestStatsServiceImpl struct {
	ITestStatsService

	LoggerService logger.ILoggerService
}

func (c *TestStatsServiceImpl) Start(testName string) {
	lg := c.LoggerService.GetFileLogger(fmt.Sprintf("stats-%v", testName), "", 1)
	ld := map[string]interface{}{}

	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		//fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
		//fmt.Printf("\tTotalAlloc = %v MiB", bToKb(m.TotalAlloc))
		//fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
		//fmt.Printf("\tNumGC = %v\n", m.NumGC)
		logger.Result(ld, fmt.Sprintf("%v,%v,%s", bToKb(m.Sys), bToKb(m.TotalAlloc), time.Duration(time.Now().UnixNano()-int64(m.LastGC))))
		logger.Print(lg, ld)

		time.Sleep(10 * time.Second)
	}
}
