package webhook

import (
	"encoding/json"
	"fmt"
	"gh-webhook/pkg/config"
	"gh-webhook/pkg/core"
	"gh-webhook/pkg/launcher"
	"gh-webhook/pkg/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
)

type GHWebhookDeliverHandler struct {
	queue        model.Queue
	wg           sync.WaitGroup
	routineId    int32
	db           *gorm.DB
	compiledExpr sync.Map
	config       *config.Config
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
	defer func() {
		if r := recover(); r != nil {
			log.Warningf("[go routine %d] event %d panic occurred: %v", routineId, ghEvent.ID, r)
		}
	}()

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

	var receiver []model.GHWebhookReceiver
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

func (h *GHWebhookDeliverHandler) handleReceiver(routineId int32, re model.GHWebhookReceiver, event model.GHWebHookEvent, payload map[string]interface{}) model.GHWebHookEventReceiverDeliver {
	receiverDeliver := model.GHWebHookEventReceiverDeliver{
		GHWebHookReceiverId: re.ID,
	}
	receiverDeliver.Delivered = false

	if len(re.Subscribes) == 0 {
		receiverDeliver.Error = fmt.Sprintf("[go routine %d] no subscribe found for receiver %d", routineId, re.ID)
		log.Warning(receiverDeliver.Error)
		return receiverDeliver
	} else if !slices.Contains(launcher.SupportedReceiverType, re.ReceiverConfig.Type) {
		receiverDeliver.Error = fmt.Sprintf("[go routine %d] unsupported receiver type %s", routineId, re.ReceiverConfig.Type)
		log.Warning(receiverDeliver.Error)
		return receiverDeliver
	}

	for _, sub := range re.Subscribes {
		if sub.Event != event.Action {
			log.Infof("[go routine %d] skip subscribe for event %s", routineId, sub.Event)
			continue
		}

		if err := sub.Matches(payload, event); err != nil {
			log.Warningf("[go routine %d] failed to match subscribe: %v", routineId, err)
			continue
		}

		receiverDeliver.Delivered = true
		receiverDeliver.Error = h.launchDelivery(routineId, re, event).Error()
		break
	}

	return receiverDeliver
}

func (h *GHWebhookDeliverHandler) launchDelivery(routineId int32, re model.GHWebhookReceiver, event model.GHWebHookEvent) error {

	launcherInst, err := launcher.NewLauncher(re.ReceiverConfig.Type)

	if err != nil {
		return err
	}

	return launcherInst.Launch(routineId, h.config, re, event)
}

func (h *GHWebhookDeliverHandler) Get(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"queue": len(h.queue),
	})
}

func (h *GHWebhookDeliverHandler) Register(c *core.GHPRContext) error {
	h.queue = model.GetQueue()
	h.wg = sync.WaitGroup{}
	h.routineId = 0
	h.db = c.Db
	h.compiledExpr = sync.Map{}
	h.config = c.Cfg
	h.Start(4)
	c.Gin.GET(fmt.Sprintf("%s/gh-webhook/handler/queue", c.Cfg.APIPrefix), h.Get)
	return nil
}
