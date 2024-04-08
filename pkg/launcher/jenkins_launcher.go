package launcher

import (
	"encoding/json"
	"fmt"
	"gh-webhook/pkg/config"
	"gh-webhook/pkg/model"
	"github.com/PaesslerAG/jsonpath"
)

type JenkinsLauncher struct {
	HttpAppLauncher
}

func (h *JenkinsLauncher) GetPayload(c config.Config, re model.GHWebHookReceiver, event model.GHWebHookEvent) ([]byte, error) {
	parameterObj, err := jsonpath.Get("$.parameter", re.ReceiverConfig.Config)

	if err != nil {
		return nil, err
	}

	parameter := parameterObj.(string)
	if len(parameter) == 0 {
		return nil, fmt.Errorf("invalid parameter")
	}

	payload := map[string]interface{}{
		"url":   fmt.Sprintf("%s/event/%d", c.APIUrl, event.ID),
		"event": event,
	}

	jenkinsPayload := map[string]interface{}{
		parameter: payload,
	}

	return json.Marshal(jenkinsPayload)
}
