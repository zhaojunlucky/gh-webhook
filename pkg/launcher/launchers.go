package launcher

import (
	"fmt"
	"gh-webhook/pkg/model"
)

func NewLauncher(receiverType string) (GHWebhookReceiverLauncher, error) {
	switch receiverType {
	case model.HTTP:
		return &HttpAppLauncher{}, nil
	case model.Jenkins:
		return &JenkinsLauncher{}, nil
	default:
		return nil, fmt.Errorf("unsupported receiver type: %s", receiverType)
	}
}
