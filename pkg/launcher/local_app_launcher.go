package launcher

import (
	"errors"
	"fmt"
	"gh-webhook/pkg/config"
	"gh-webhook/pkg/model"
	"github.com/PaesslerAG/jsonpath"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

type LocalAppLauncher struct {
}

func (h *LocalAppLauncher) IsAppExists(name string) (bool, error) {
	_, err := exec.LookPath(name)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return false, nil // Application not found
		}
		return false, err // Other error
	}
	return true, nil // Application found
}

func (h *LocalAppLauncher) Launch(routineId int32, config config.Config, re model.GHWebHookReceiver, event model.GHWebHookEvent) error {
	url := fmt.Sprintf("%s/event/%d", config.APIUrl, event.ID)
	receiverCfg := re.ReceiverConfig.Config

	appObj, err := jsonpath.Get("$.app", receiverCfg)

	if err != nil {
		return err
	}

	app := appObj.(string)
	if len(app) == 0 {
		return fmt.Errorf("invalid app")
	}

	ok, err := h.IsAppExists(app)

	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("app %s doesn't exist", app)
	}

	payloadArgObj, err := jsonpath.Get("$.payloadArg", receiverCfg)
	if err != nil {
		return err
	}

	payloadArg := payloadArgObj.(string)

	if len(payloadArg) == 0 {
		return fmt.Errorf("invalid payloadArg")
	}

	otherArgsObj, err := jsonpath.Get("$.args", receiverCfg)

	if err != nil {
		return err
	}

	otherArgs := otherArgsObj.([]string)

	args := append([]string{}, payloadArg, url)
	args = append(args, otherArgs...)

	cmd := exec.Command(app, args...)
	err = cmd.Start()
	if err != nil {
		return err
	}
	log.Infof("[go routine %d] launched app %s, process id %d", routineId, app, cmd.Process.Pid)
	return nil
}
