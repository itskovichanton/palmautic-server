package frontend

import (
	"encoding/json"
	"github.com/itskovichanton/echo-http"
	entities2 "github.com/itskovichanton/server/pkg/server/entities"
	"github.com/itskovichanton/server/pkg/server/pipeline"
	"io"
	"salespalm/server/app/backend/tests"
)

type StartSeqTestAction struct {
	pipeline.BaseActionImpl

	TestService tests.ITestService
}

func (c *StartSeqTestAction) Run(arg interface{}) (interface{}, error) {
	cp := arg.(*entities2.CallParams)
	bodyBytes, err := io.ReadAll(cp.Request.(echo.Context).Request().Body)
	if err != nil {
		return nil, err
	}
	var settings tests.SeqTestSettings
	err = json.Unmarshal(bodyBytes, &settings)
	if err != nil {
		return nil, err
	}
	c.TestService.StartSequencesTest(&settings)
	return "started", nil
}
