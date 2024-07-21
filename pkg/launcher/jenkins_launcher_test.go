package launcher

import (
	"fmt"
	"gh-webhook/pkg/config"
	"gh-webhook/pkg/model"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestJenkinsLauncher_GetPayload(t *testing.T) {
	launcher := JenkinsLauncher{}

	cfg := config.Config{
		APIUrl: "http://localhost:8080",
	}

	github := model.GitHub{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			DeletedAt: gorm.DeletedAt{},
			UpdatedAt: time.Now(),
		},
		Web:  "https://github.com",
		API:  "https://api.github.com/v3",
		Name: "github",
	}

	re := model.GHWebhookReceiver{
		Model:    gorm.Model{},
		Name:     "test",
		GitHubId: github.ID,
		GitHub:   github,
		ReceiverConfig: model.GHWebhookReceiverConfig{
			Type:      "jenkins",
			URL:       "http://localhost:8080",
			Auth:      "none",
			Parameter: "payload",
		},
		Subscribes: nil,
	}

	event := model.GHWebhookEvent{
		Model: gorm.Model{
			ID:        2,
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
			DeletedAt: gorm.DeletedAt{},
		},
		HookMeta: map[string]string{
			"Context-Type": "application/json",
		},
		Payload:   "{\"event\": \"pull_request\", \"action\":\"open\",}",
		Event:     "pull_request",
		Action:    "open",
		OrgRepo:   "zhaojunlucky/veda",
		HookId:    "X-GitHub-Hook-ID",
		PayloadId: "X-GitHub-Delivery",
		GitHubId:  github.ID,
		GitHub:    github,
	}

	deliver := model.GHWebhookEventReceiverDeliver{
		Model: gorm.Model{
			ID:        4,
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
			DeletedAt: gorm.DeletedAt{},
		},
		GHWebhookReceiverId:     4,
		GHWebhookEventDeliverID: 5,
		GHWebhookEventDeliver: model.GHWebhookEventDeliver{
			Model: gorm.Model{
				ID:        5,
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
				DeletedAt: gorm.DeletedAt{},
			},
			GHWebhookEventId:   2,
			GHWebhookEvent:     event,
			GHWebhookReceivers: nil,
			Error:              "",
			Delivered:          false,
		},
		Delivered: false,
		Error:     "",
		Ack:       "",
	}
	payload, err := launcher.GetPayload(&cfg, re, event, deliver)
	if err != nil {
		t.Fatal(err)
	}

	data := string(payload)
	fmt.Println(data)

}
