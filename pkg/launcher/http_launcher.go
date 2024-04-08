package launcher

import (
	"encoding/json"
	"fmt"
	"gh-webhook/pkg/config"
	"gh-webhook/pkg/model"
	"github.com/PaesslerAG/jsonpath"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"slices"
	"strings"
)

type HttpAppLauncher struct {
}

func (h *HttpAppLauncher) Launch(routineId int32, config config.Config, re model.GHWebHookReceiver, event model.GHWebHookEvent) error {

	str, err := h.GetPayload(config, re, event)
	if err != nil {
		return err
	}

	receiverCfg := re.ReceiverConfig.Config

	urlObj, err := jsonpath.Get("$.url", receiverCfg)
	if err != nil {
		return err
	}
	url := urlObj.(string)
	if len(url) == 0 {
		return fmt.Errorf("invalid url")
	}

	authObj, err := jsonpath.Get("$.auth", receiverCfg)
	auth := "none"
	if err != nil {
		log.Warnf("failed to get auth: %v", err)
	} else {
		auth = authObj.(string)
	}

	if !slices.Contains(SupportedAuthType, auth) {
		return fmt.Errorf("unsupported auth type %s", auth)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(str)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if auth == "basic" || auth == "token" {
		userObj, err := jsonpath.Get("$.username", receiverCfg)
		if err != nil {
			return err
		}
		username := userObj.(string)

		passwordObj, err := jsonpath.Get("$.password", receiverCfg)

		if err != nil {
			return err
		}
		password := passwordObj.(string)

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

func (h *HttpAppLauncher) GetPayload(c config.Config, re model.GHWebHookReceiver, event model.GHWebHookEvent) ([]byte, error) {
	payload := map[string]interface{}{
		"url":   fmt.Sprintf("%s/event/%d", c.APIUrl, event.ID),
		"event": event,
	}

	return json.Marshal(payload)
}
