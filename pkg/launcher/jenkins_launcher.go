package launcher

import (
	"encoding/json"
	"fmt"
	"gh-webhook/pkg/config"
	"gh-webhook/pkg/model"
)

type JenkinsLauncher struct {
	HttpAppLauncher
}

func (h *JenkinsLauncher) GetPayload(c *config.Config, re model.GHWebhookReceiver, event model.GHWebHookEvent) ([]byte, error) {

	parameter := re.ReceiverConfig.Parameter
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
