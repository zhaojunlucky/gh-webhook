package launcher

import (
	"fmt"
	"gh-webhook/pkg/config"
	"gh-webhook/pkg/model"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHttpAppLauncher_GetPayload(t *testing.T) {
	launcher := HttpAppLauncher{}

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
		Model:          gorm.Model{},
		Name:           "test",
		GitHubId:       github.ID,
		GitHub:         github,
		ReceiverConfig: model.GHWebhookReceiverConfig{},
		Subscribes:     nil,
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

func Test_Launch(t *testing.T) {
	launcher := HttpAppLauncher{}

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

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close() // Close the server after the test

	re := model.GHWebhookReceiver{
		Model:    gorm.Model{},
		Name:     "test",
		GitHubId: github.ID,
		GitHub:   github,
		ReceiverConfig: model.GHWebhookReceiverConfig{
			Type: "http",
			URL:  ts.URL,
			Auth: "none",
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

	err := launcher.Launch(1, &cfg, re, event, deliver)
	if err != nil {
		t.Fatal(err)
	}
}

func handler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(200)
	fmt.Fprintf(writer, "Hello, client!")
}
