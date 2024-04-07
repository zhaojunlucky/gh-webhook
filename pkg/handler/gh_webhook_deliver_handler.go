package handler

import (
	"encoding/json"
	"fmt"
	"gh-webhook/pkg/model"
	"github.com/PaesslerAG/jsonpath"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
)

type Queue chan model.GHWebHookEvent

type GHWebhookDeliverHandler struct {
	queue        Queue
	wg           sync.WaitGroup
	routineId    int32
	db           *gorm.DB
	compiledExpr sync.Map
}

func (h *GHWebhookDeliverHandler) Start(processors int) {
	h.wg.Add(processors)

	for i := 0; i < processors; i++ {
		go h.handleWebHook()
	}
}

func (h *GHWebhookDeliverHandler) handleWebHook() {
	routineId := atomic.AddInt32(&h.routineId, 1)

	log.Infof("[go routine %d] started", routineId)
	defer h.wg.Done()
	for {
		ghEvent, ok := <-h.queue // Receive with value and ok result
		if !ok {
			log.Warningf("Channel closed, go routine %d exited", routineId)
			return
		}
		log.Infof("[go routine %d] received web hook event %s action %s with payload %d", routineId,
			ghEvent.Event, ghEvent.Action, ghEvent.ID)
		h.handle(routineId, ghEvent)

	}
}

func (h *GHWebhookDeliverHandler) Close() error {
	h.wg.Wait()
	return nil
}

func (h *GHWebhookDeliverHandler) handle(routineId int32, ghEvent model.GHWebHookEvent) {
	receiverLog := model.GHWebhookEventDelivers{
		GHWebHookEventId: ghEvent.ID,
		GHWebHookEvent:   ghEvent,
	}

	defer func() {
		r := h.db.Create(&receiverLog)
		if r.Error != nil {
			log.Errorf("[go routine %d] failed to create receiver log: %v", routineId, r.Error)
		}
	}()

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(ghEvent.Payload), &payload); err != nil {
		log.Errorf("[go routine %d] failed to parse payload as json: %v", routineId, err)
		receiverLog.Delivered = false
		receiverLog.Error = fmt.Sprintf("failed to parse payload as json: %v", err)
		return
	}

	var receiver []model.GHWebHookReceiver
	r := h.db.Where("github_id = ?", ghEvent.GitHubId).Find(&receiver)
	if r.Error != nil {
		log.Errorf("[go routine %d] failed to find receiver: %v", routineId, r.Error)
		receiverLog.Delivered = false
		receiverLog.Error = fmt.Sprintf("no receivers found: %v", r.Error)
		return
	} else if len(receiver) == 0 {
		log.Infof("[go routine %d] no receiver found", routineId)
		receiverLog.Delivered = false
		receiverLog.Error = "no receivers found"

		return
	}
	receiverLog.Delivered = true
	ids := make([]string, len(receiver))
	for _, re := range receiver {
		ids = append(ids, fmt.Sprintf("%d", re.ID))
	}
	log.Infof("[go routine %d] found receivers %s", routineId, strings.Join(ids, ", "))

	for _, re := range receiver {
		receiverLog.GHWebHookReceivers = append(receiverLog.GHWebHookReceivers, h.handleReceiver(routineId, re, ghEvent, payload))
	}
}

func (h *GHWebhookDeliverHandler) handleReceiver(routineId int32, re model.GHWebHookReceiver, event model.GHWebHookEvent, payload map[string]interface{}) model.GHWebHookEventReceiverDeliver {
	receiverDeliver := model.GHWebHookEventReceiverDeliver{
		GHWebHookReceiverId: re.ID,
	}
	receiverDeliver.Delivered = false

	if len(re.Subscribes) == 0 {
		receiverDeliver.Error = fmt.Sprintf("[go routine %d] no subscribe found for receiver %d", routineId, re.ID)
		log.Warning(receiverDeliver.Error)
		return receiverDeliver
	} else if !slices.Contains(model.SupportedReceiverType, re.ReceiverConfig.Type) {
		receiverDeliver.Error = fmt.Sprintf("[go routine %d] unsupported receiver type %s", routineId, re.ReceiverConfig.Type)
		log.Warning(receiverDeliver.Error)
		return receiverDeliver
	}

	for _, sub := range re.Subscribes {
		if sub.Event != event.Action {
			log.Infof("[go routine %d] skip subscribe for event %s", routineId, sub.Event)
			continue
		}

		if len(sub.Actions) > 0 && !slices.Contains(sub.Actions, event.Action) {
			log.Infof("[go routine %d] skip subscribe for action %s", routineId, event.Action)
			continue
		}

		if err := h.matchOrgRepo(routineId, sub.OrgRepo, event); err != nil {
			receiverDeliver.Delivered = true
			receiverDeliver.Error = err.Error()
			continue
		}

		if err := h.match(routineId, sub, event, payload); err != nil {
			receiverDeliver.Delivered = true
			receiverDeliver.Error = err.Error()
			continue
		}

		if err := h.matchExpr(routineId, sub, event); err != nil {
			receiverDeliver.Delivered = true
			receiverDeliver.Error = err.Error()
			continue
		}

		receiverDeliver.Delivered = true
		receiverDeliver.Error = h.launchDelivery(routineId, re, event)
		break
	}

	return receiverDeliver
}

func (h *GHWebhookDeliverHandler) launchDelivery(routineId int32, re model.GHWebHookReceiver, event model.GHWebHookEvent) string {
	if re.ReceiverConfig.Type != "local" {
		return h.launchLocal(routineId, re, event)
	} else if re.ReceiverConfig.Type != "http" {
		return h.launchHttp(routineId, re, event)
	} else {
		return fmt.Sprintf("invalid receiver type %s", re.ReceiverConfig.Type)
	}
}

func (h *GHWebhookDeliverHandler) launchLocal(routineId int32, re model.GHWebHookReceiver, event model.GHWebHookEvent) string {
	return ""
}

func (h *GHWebhookDeliverHandler) launchHttp(routineId int32, re model.GHWebHookReceiver, event model.GHWebHookEvent) string {
	return ""
}

func (h *GHWebhookDeliverHandler) matchOrgRepo(routineId int32, orgRepo string, event model.GHWebHookEvent) error {
	if len(orgRepo) > 0 {
		matched, err := regexp.MatchString(orgRepo, event.OrgRepo)
		if err != nil {
			log.Errorf("failed to match org repo %s: %v", orgRepo, err)
			return err
		}
		if !matched {
			log.Infof("[go routine %d] skip subscribe for org repo %s", routineId, orgRepo)
			return fmt.Errorf("[go routine %d] skip subscribe for org repo %s", routineId, orgRepo)
		} else {
			log.Infof("[go routine %d] matched org repo %s", routineId, orgRepo)
			return nil
		}
	}
	return nil
}

func (h *GHWebhookDeliverHandler) matchExpr(routineId int32, sub model.GHWebHookSubscribe, event model.GHWebHookEvent) error {
	if len(sub.Expr) > 0 {
		prog, ok := h.compiledExpr.Load(sub.Expr)
		var program *vm.Program
		var err error
		if !ok {
			program, err = expr.Compile(sub.Expr, expr.AsBool())
			if err != nil {
				log.Errorf("failed to compile expr %s: %v", sub.Expr, err)
				return err
			}
			h.compiledExpr.Store(sub.Expr, program)
		} else {
			program = prog.(*vm.Program)
		}
		env := map[string]interface{}{
			"event": event,
		}
		output, err := expr.Run(program, env)
		if err != nil {
			log.Errorf("failed to run expr %s: %v", sub.Expr, err)
		}
		if output.(bool) {
			log.Infof("[go routine %d] matched expr %s", routineId, sub.Expr)
		}
	}
	return nil
}

func (h *GHWebhookDeliverHandler) match(routineId int32, sub model.GHWebHookSubscribe, event model.GHWebHookEvent, payload map[string]interface{}) error {
	switch sub.Event {
	case PullRequest:
		return h.handlerPullRequest(routineId, sub, event, payload)
	case Push:
		return h.handlerPush(routineId, sub, event, payload)
	case IssueComment:
		return h.handlerIssueComment(routineId, sub, event, payload)
	default:
		return nil
	}
}

func (h *GHWebhookDeliverHandler) handlerPullRequest(routineId int32, sub model.GHWebHookSubscribe, event model.GHWebHookEvent, payload map[string]interface{}) error {
	if len(sub.PullRequest.AllowedBaseBranches) <= 0 && len(sub.PullRequest.DisallowedBaseBranches) <= 0 {
		return nil
	}

	baseBranchObj, err := jsonpath.Get(fmt.Sprintf("%s.base.ref", PullRequest), payload)
	if err != nil {
		return err
	}

	baseBranch := baseBranchObj.(string)

	if len(sub.PullRequest.DisallowedBaseBranches) > 0 {
		for _, disallowedBaseBranch := range sub.PullRequest.DisallowedBaseBranches {
			matched, err := regexp.MatchString(disallowedBaseBranch, baseBranch)
			if err != nil {
				return err
			}
			if matched {
				return fmt.Errorf("[go routine %d] disallowed base branch %s", routineId, baseBranch)
			}
			continue
		}
	}

	if len(sub.PullRequest.AllowedBaseBranches) > 0 {
		for _, allowedBaseBranch := range sub.PullRequest.AllowedBaseBranches {
			matched, err := regexp.MatchString(allowedBaseBranch, baseBranch)
			if err != nil {
				return err
			}
			if matched {
				return nil
			}
		}
		return fmt.Errorf("[go routine %d] allowed base branch doesn't match", routineId)
	}

	return nil
}

func (h *GHWebhookDeliverHandler) handlerPush(routineId int32, sub model.GHWebHookSubscribe, event model.GHWebHookEvent, payload map[string]interface{}) error {
	if len(sub.Push.AllowedPushRefs) <= 0 {
		return nil
	}

	pushRefObj, err := jsonpath.Get("ref", payload)

	if err != nil {
		return err
	}

	pushRef := pushRefObj.(string)

	for _, allowedPushRef := range sub.Push.AllowedPushRefs {
		matched, err := regexp.MatchString(allowedPushRef, pushRef)
		if err != nil {
			return err
		}
		if matched {
			return nil
		}
	}
	return fmt.Errorf("[go routine %d] allowed push ref doesn't match", routineId)
}

func (h *GHWebhookDeliverHandler) handlerIssueComment(id int32, sub model.GHWebHookSubscribe, event model.GHWebHookEvent, payload map[string]interface{}) error {
	if len(sub.IssueComment.AllowedComments) <= 0 {
		return nil
	}

	commentObj, err := jsonpath.Get("comment.body", payload)

	if err != nil {
		return err
	}

	comment := commentObj.(string)

	for _, allowedComment := range sub.IssueComment.AllowedComments {
		matched, err := regexp.MatchString(allowedComment, comment)
		if err != nil {
			return err
		}
		if matched {
			return nil
		}
	}
	return fmt.Errorf("[go routine %d] allowed comment doesn't match", id)
}
