package model

import (
	"gorm.io/gorm"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_GHWebhookFieldVal(t *testing.T) {
	fieldVal := GHWebhookFieldVal{
		Value: 123,
		Type:  reflect.TypeOf(123),
	}
	if !fieldVal.IsNumeric() {
		t.Error("should be numeric")
	}

	if fieldVal.GetAsString() != "123" {
		t.Error("should be 123")
	}

	fieldVal = GHWebhookFieldVal{
		Value: 123.9,
		Type:  reflect.TypeOf(123.9),
	}
	if !fieldVal.IsNumeric() {
		t.Error("should be numeric")
	}
	if !strings.HasPrefix(fieldVal.GetAsString(), "123.9") {
		t.Error("should be 123.9")
	}

}

func Test_GHWebhookField_Invalid(t *testing.T) {

	field := GHWebhookField{
		PositiveMatches: nil,
		NegativeMatches: nil,
		Child:           nil,
		Expr:            "",
	}

	if err := field.IsValid(); err == nil {
		t.Error("should be failed")
	}
}

func Test_GHWebhookField_Valid(t *testing.T) {
	field := GHWebhookField{
		PositiveMatches: nil,
		NegativeMatches: nil,
		Child:           nil,
		Expr:            "sdsd",
	}
	if err := field.IsValid(); err != nil {
		t.Error(err)
	}
}

func Test_GHWebhookField_Matches_No_Child(t *testing.T) {
	field := GHWebhookField{
		PositiveMatches: []string{"^integration/.+$"},
		NegativeMatches: []string{"^usr/.+$"},
		Child:           nil,
		Expr:            "",
	}
	if err := field.IsValid(); err != nil {
		t.Error(err)
	}

	payload := map[string]interface{}{
		"branch": "integration/main",
	}
	ghEvent := GHWebhookEvent{
		Model: gorm.Model{
			ID:        0,
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
			DeletedAt: gorm.DeletedAt{},
		},
	}

	if err := field.Matches(payload, ghEvent, "branch"); err != nil {
		t.Error(err)
	}

	payload = map[string]interface{}{
		"branch": "usr/main",
	}

	if err := field.Matches(payload, ghEvent, "branch"); err == nil {
		t.Error("should be failed")
	}
}

func Test_GHWebhookField_Matches_Expr(t *testing.T) {
	field := GHWebhookField{
		PositiveMatches: []string{"^integration/.+$"},
		NegativeMatches: []string{"^usr/.+$"},
		Child:           nil,
		Expr:            "cur != \"integration/main\"",
	}
	if err := field.IsValid(); err != nil {
		t.Error(err)
	}

	payload := map[string]interface{}{
		"branch": "integration/main",
	}
	ghEvent := GHWebhookEvent{
		Model: gorm.Model{
			ID:        0,
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
			DeletedAt: gorm.DeletedAt{},
		},
	}

	err := field.Matches(payload, ghEvent, "branch")
	if err == nil {
		t.Error("should be failed")
	}

}
