package model

import (
	"strings"
	"testing"
)

func TestGHWebhookReceiverConfig_InValidAuth(t *testing.T) {
	cfg := GHWebhookReceiverConfig{
		Auth: "sd",
	}

	err := cfg.IsValid()

	if err == nil || !strings.Contains(err.Error(), "invalid auth type") {
		t.Error("expected error")
	}

}

func TestGHWebhookReceiverConfig_InValidType(t *testing.T) {
	cfg := GHWebhookReceiverConfig{
		Auth: NoneAuth,
		Type: "sdf",
	}

	err := cfg.IsValid()

	if err == nil || !strings.Contains(err.Error(), "invalid receiver type") {
		t.Error("expected error")
	}
}

func TestGHWebhookReceiverConfig_InValidParameter(t *testing.T) {
	cfg := GHWebhookReceiverConfig{
		Auth:      NoneAuth,
		Type:      Jenkins,
		Parameter: " ",
	}

	err := cfg.IsValid()

	if err == nil || !strings.Contains(err.Error(), "invalid parameter") {
		t.Error("expected error")
	}
}

func TestGHWebhookReceiverConfig_InValidUser(t *testing.T) {
	cfg := GHWebhookReceiverConfig{
		Auth:      BasicAuth,
		Type:      Jenkins,
		Parameter: "payload",
	}

	err := cfg.IsValid()

	if err == nil || !strings.Contains(err.Error(), "username/token header or password/token value is empty") {
		t.Error("expected error")
	}
}

func TestGHWebhookReceiverConfig_IsValid(t *testing.T) {
	cfg := GHWebhookReceiverConfig{
		Auth:      BasicAuth,
		Type:      Jenkins,
		Parameter: "payload",
		Username:  "aaa",
		Password:  "sdsd",
	}

	err := cfg.IsValid()
	if err != nil {
		t.Fatal(err)
	}
}
