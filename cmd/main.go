package main

import (
	app2 "github.com/itskovichanton/core/pkg/core/app"
	"salespalm/server/app"
)

func main() {

	di := &app.DI{}
	di.InitDI()

	var outerAppRunner app2.IAppRunner
	err := di.Container.Invoke(func(app app2.IAppRunner) {
		outerAppRunner = app
	})

	if err != nil {
		panic(err)
	}

	runningErr := outerAppRunner.Run()
	if runningErr != nil {
		panic(runningErr)
	}
}
