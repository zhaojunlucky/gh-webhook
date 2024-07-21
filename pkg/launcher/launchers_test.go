package launcher

import (
	"gh-webhook/pkg/model"
	"reflect"
	"testing"
)

func TestNewLauncher(t *testing.T) {
	_, err := NewLauncher("unknown")
	if err == nil {
		t.Error("new launcher should return error")
	}
}

func TestNewLauncherHttp(t *testing.T) {
	launcher, err := NewLauncher(model.HTTP)
	if err != nil {
		t.Error(err)
	}
	lType := reflect.TypeOf(launcher)
	if lType.Elem().Name() != "HttpAppLauncher" {
		t.Error("launcher type should be HttpAppLauncher")
	}
}

func TestNewLauncherJenkins(t *testing.T) {
	launcher, err := NewLauncher(model.Jenkins)
	if err != nil {
		t.Error(err)
	}
	lType := reflect.TypeOf(launcher)
	if lType.Elem().Name() != "JenkinsLauncher" {
		t.Error("launcher type should be JenkinsLauncher")
	}
}
