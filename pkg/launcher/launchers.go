package launcher

import (
	"fmt"
)

func NewLauncher(receiverType string) (GHWebhookReceiverLauncher, error) {
	switch receiverType {
	case Http:
		return &HttpAppLauncher{}, nil
	case Jenkins:
		return &JenkinsLauncher{}, nil
	default:
		return nil, fmt.Errorf("unsupported receiver type: %s", receiverType)
	}
}
