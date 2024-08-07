package webhook

import (
	"encoding/json"
	"fmt"
	"gh-webhook/pkg/core"
	"gh-webhook/pkg/model"
	"github.com/PaesslerAG/jsonpath"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io"
	"net/http"
)
import "github.com/gin-gonic/gin"

// path: gh-webhook
type GHWebhookHandler struct {
	db    *gorm.DB
	queue model.Queue
}

// Post receive webhook post event from github
func (h *GHWebhookHandler) Post(c *gin.Context) {
	if !h.validate(c) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Failed to validate webhook"})
		return
	}
	var github model.GitHub
	r := h.db.First(&github, "api = ?", c.Request.Host)
	if r.Error != nil {
		log.Errorf("failed to find github server: %v", r.Error)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "Failed to find github server"})
	}

	// add validate source

	ghHookEvent := model.GHWebhookEvent{
		GitHub:   github,
		GitHubId: github.ID,
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Errorf("failed to read payload from wehbook: %v", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "Failed to read payload"})
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Errorf("failed to unmarshal payload: %v", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "Failed to unmarshal payload as map[string]interface{}"})
	}
	action, err := jsonpath.Get("$.action", payload)
	if err != nil {
		log.Errorf("failed to get action from payload: %v", err)
	} else {
		ghHookEvent.Action = action.(string)
	}

	ghHeaders := make(map[string]string)
	ghHeaders["X-GitHub-Hook-ID"] = c.Request.Header.Get("X-GitHub-Hook-ID")
	ghHookEvent.HookId = c.Request.Header.Get("X-GitHub-Hook-ID")

	ghHeaders["X-GitHub-Event"] = c.Request.Header.Get("X-GitHub-Event")
	ghHookEvent.Event = c.Request.Header.Get("X-GitHub-Event")

	ghHeaders["X-GitHub-Delivery"] = c.Request.Header.Get("X-GitHub-Delivery")
	ghHookEvent.PayloadId = c.Request.Header.Get("X-GitHub-Delivery")

	ghHeaders["X-GitHub-Hook-Installation-Target-Type"] = c.Request.Header.Get("X-GitHub-Hook-Installation-Target-Type")
	ghHeaders["X-GitHub-Hook-Installation-Target-ID"] = c.Request.Header.Get("X-GitHub-Hook-Installation-Target-ID")
	// add X-Hub-Signature-256
	ghHookEvent.HookMeta = ghHeaders
	ghHookEvent.Payload = string(body)

	r = h.db.Create(&ghHookEvent)
	if r.Error != nil {
		log.Errorf("failed to create webhook event: %v", r.Error)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"status": "Failed to create webhook event"})
	}
	// push to queue
	go h.push(ghHookEvent)
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (h *GHWebhookHandler) push(event model.GHWebhookEvent) {
	h.queue <- event
}

func (h *GHWebhookHandler) validate(c *gin.Context) bool {
	return true
}

func (h *GHWebhookHandler) Register(c *core.GHPRContext) error {
	h.db = c.Db
	h.queue = model.GetQueue()
	c.Gin.POST(fmt.Sprintf("%s/gh-webhook", c.Cfg.APIPrefix), h.Post)
	return nil
}
