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

func (h *JenkinsLauncher) GetPayload(c *config.Config, re model.GHWebhookReceiver, event model.GHWebhookEvent, receiverLog model.GHWebhookEventDeliver) ([]byte, error) {

	parameter := re.ReceiverConfig.Parameter
	if len(parameter) == 0 {
		return nil, fmt.Errorf("invalid parameter")
	}

	payload := map[string]interface{}{
		"url":             fmt.Sprintf("%s/event/%d", c.APIUrl, event.ID),
		"event":           event,
		"eventDeliverUrl": fmt.Sprintf("%s/gh-webhook-event-deliver/%d", c.APIUrl, receiverLog.ID),
	}

	jenkinsPayload := map[string]interface{}{
		parameter: payload,
	}

	return json.Marshal(jenkinsPayload)
}
