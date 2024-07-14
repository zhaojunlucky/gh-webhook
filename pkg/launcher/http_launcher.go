package launcher

import (
	"encoding/json"
	"fmt"
	"gh-webhook/pkg/config"
	"gh-webhook/pkg/model"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"slices"
	"strings"
)

type HttpAppLauncher struct {
}

func (h *HttpAppLauncher) Launch(routineId int32, config *config.Config, re model.GHWebhookReceiver, event model.GHWebhookEvent,
	receiverDeliver model.GHWebhookEventReceiverDeliver) error {

	str, err := h.GetPayload(config, re, event, receiverDeliver)
	if err != nil {
		return err
	}

	url := re.ReceiverConfig.URL
	if len(url) == 0 {
		return fmt.Errorf("invalid url")
	}

	auth := re.ReceiverConfig.Auth
	if !slices.Contains(SupportedAuthType, auth) {
		return fmt.Errorf("unsupported auth type %s", auth)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(str)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if auth == "basic" || auth == "token" {
		username := re.ReceiverConfig.Username

		password := re.ReceiverConfig.Password

		if len(username) == 0 || len(password) == 0 {
			return fmt.Errorf("username/token header or password/token value is empty")
		}

		if auth == "basic" {
			req.SetBasicAuth(username, password)
		} else {
			req.Header.Add(username, password)
		}
	}
	client := http.DefaultClient

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to send request: %s", resp.Status)
	}
	body := "unknown"
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[go routine %d] failed to read response body: %v", routineId, err)
	} else {
		body = string(data)
	}

	log.Infof("succeed to send request: %s, body: %s", resp.Status, body)
	return nil
}

func (h *HttpAppLauncher) GetPayload(c *config.Config, re model.GHWebhookReceiver, event model.GHWebhookEvent, receiverDeliver model.GHWebhookEventReceiverDeliver) ([]byte, error) {
	payload := map[string]interface{}{
		"url":                fmt.Sprintf("%s/gh-webhook-event/%d", c.APIUrl, event.ID),
		"event":              event,
		"eventDeliverAckUrl": fmt.Sprintf("%s/gh-webhook-receiver-deliver/%d/ack", c.APIUrl, receiverDeliver.ID),
	}

	return json.Marshal(payload)
}
