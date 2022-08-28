package main

import (
	app2 "bitbucket.org/itskovich/core/pkg/core/app"
	"palm/app"
)

func main() {

	di := &app.DI{}
	di.InitDI()

	var outerApp app2.IApp
	err := di.Container.Invoke(func(app app2.IApp) {
		outerApp = app
	})

	if err != nil {
		panic(err)
	}

	runningErr := outerApp.Run()
	if runningErr != nil {
		panic(runningErr)
	}
}
