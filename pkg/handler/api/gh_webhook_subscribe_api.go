package api

import (
	"errors"
	"fmt"
	"gh-webhook/pkg/core"
	"gh-webhook/pkg/model"
	"github.com/dranikpg/dto-mapper"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

// GHWebhookSubscribeAPIHandler path: gh-webhook-receiver/<receiver id>/subscribe
type GHWebhookSubscribeAPIHandler struct {
	db *gorm.DB
}

type GHWebhookFieldCreateDTO struct {
	PositiveMatches []string                           `json:"positiveMatches"`
	NegativeMatches []string                           `json:"negativeMatches"`
	Child           map[string]GHWebhookFieldCreateDTO `json:"child"`
	Expr            string                             `json:"expr"`
}

type GHWebhookSubscribeCreateDTO struct {
	GHWebHookReceiverID uint
	Event               string `json:"event"`

	Filters map[string]GHWebhookFieldCreateDTO `json:"filters"`
}

type GHWebhookSubscribeSearchDTO struct {
	ID                  uint   `json:"id" rsql:"id,filter,sort"`
	GHWebHookReceiverID uint   `json:"recieverId" rsql:"recieverId,filter,sort"`
	Event               string `json:"event"`

	Filters map[string]GHWebhookFieldSearchDTO `json:"filters"`
}

type GHWebhookFieldSearchDTO struct {
	PositiveMatches []string                           `json:"positiveMatches"`
	NegativeMatches []string                           `json:"negativeMatches"`
	Child           map[string]GHWebhookFieldSearchDTO `json:"child"`
	Expr            string                             `json:"expr"`
}

type GHWebhookSubscribeUpdateDTO struct {
	Event string `json:"event"`

	Filters map[string]GHWebhookFieldUpdateDTO `json:"filters"`
}

type GHWebhookFieldUpdateDTO struct {
	PositiveMatches []string                           `json:"positiveMatches"`
	NegativeMatches []string                           `json:"negativeMatches"`
	Child           map[string]GHWebhookFieldUpdateDTO `json:"child"`
	Expr            string                             `json:"expr"`
}

func (h *GHWebhookSubscribeAPIHandler) Register(c *core.GHPRContext) error {
	h.db = c.Db
	c.Gin.POST(fmt.Sprintf("%s/gh-webhook-receiver/:id/subscribe", c.Cfg.APIPrefix), h.Post)
	c.Gin.PATCH(fmt.Sprintf("%s/gh-webhook-receiver/:id/subscribe/:cId", c.Cfg.APIPrefix), h.Update)
	c.Gin.GET(fmt.Sprintf("%s/gh-webhook-receiver/:id/subscribe/:cId", c.Cfg.APIPrefix), h.Get)
	c.Gin.DELETE(fmt.Sprintf("%s/gh-webhook-receiver/:id/subscribe/:cId", c.Cfg.APIPrefix), h.Delete)
	c.Gin.GET(fmt.Sprintf("%s/gh-webhook-receiver/:id/subscribe", c.Cfg.APIPrefix), h.List)
	return nil
}

func (h *GHWebhookSubscribeAPIHandler) Post(c *gin.Context) {
	pId, err := core.UIntParam(c, "id")
	if err != nil {
		log.Errorf("failed to convert pId: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}

	var receiver model.GHWebhookReceiver

	db := h.db.First(&receiver, "id = ?", pId)
	if db.Error != nil {
		log.Errorf("failed to find webhook receiver: %v", db.Error)
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, model.NewErrorMsgDTO(http.StatusText(http.StatusNotFound)))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}

	var createDto GHWebhookSubscribeCreateDTO
	if err := c.ShouldBindJSON(&createDto); err != nil {
		log.Errorf("failed to bind json: %v", err)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(err))
		return
	}
	mapper := dto.Mapper{}
	var filters map[string]model.GHWebhookField
	err = mapper.Map(&filters, createDto.Filters)
	if err != nil {
		log.Errorf("failed to bind json: %v", err)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(err))
		return
	}

	sub := model.GHWebHookSubscribe{
		GHWebhookReceiverID: pId,
		GHWebhookReceiver:   receiver,
		Event:               createDto.Event,
		Filters:             filters,
	}

	if err = sub.IsValid(); err != nil {
		log.Errorf("invalid request: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}

	h.db.Save(&sub)
	if db.Error != nil {
		log.Errorf("failed to save webhook receiver: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}
	c.JSON(http.StatusCreated, model.NewIDResponse(sub.ID))
}

func (h *GHWebhookSubscribeAPIHandler) Get(c *gin.Context) {
	cId := core.GetPathVarUInt(c, "cId")
	if cId == nil {
		return
	}

	pId := core.GetPathVarUInt(c, "id")
	if pId == nil {
		return
	}

	receiver := model.GHWebHookSubscribe{}
	db := h.db.First(&receiver, "id = ?", *cId)
	if db.Error != nil {
		log.Errorf("failed to find webhook receiver: %v", db.Error)
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, model.NewErrorMsgDTO(http.StatusText(http.StatusBadRequest)))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}

	sub := model.GHWebHookSubscribe{}

	db = h.db.First(&sub, "id = ?", receiver.ID)
	if db.Error != nil {
		log.Errorf("failed to find webhook receiver subscribe: %v", db.Error)
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, model.NewErrorMsgDTO(http.StatusText(http.StatusNotFound)))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}

	if sub.GHWebhookReceiverID != *pId {
		log.Errorf("webhook receiver subscribe not found")
		c.JSON(http.StatusNotFound, model.NewErrorMsgDTO(http.StatusText(http.StatusNotFound)))
		return
	}
	mapper := dto.Mapper{}
	to := GHWebhookSubscribeSearchDTO{}
	err := mapper.Map(&to, sub)
	if err != nil {
		log.Errorf("failed to map: %v", err)
		c.JSON(http.StatusInternalServerError, model.NewErrorMsgDTOFromErr(err))
		return
	}
	c.JSON(http.StatusOK, to)
}

func (h *GHWebhookSubscribeAPIHandler) Update(c *gin.Context) {
	pId, err := core.UIntParam(c, "pId")
	if err != nil {
		log.Errorf("failed to convert pId: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}

	receiver := model.GHWebHookSubscribe{}
	db := h.db.First(&receiver, "id = ?", pId)
	if db.Error != nil {
		log.Errorf("failed to find webhook receiver: %v", db.Error)
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, model.NewErrorMsgDTO(http.StatusText(http.StatusBadRequest)))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}

	sub := model.GHWebHookSubscribe{}

	db = h.db.First(&sub, "id = ?", receiver.ID)
	if db.Error != nil {
		log.Errorf("failed to find webhook receiver subscribe: %v", db.Error)
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, model.NewErrorMsgDTO(http.StatusText(http.StatusNotFound)))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}

	if sub.GHWebhookReceiverID != pId {
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTO("cannot update receiver"))
		return
	}

	var updateDto GHWebhookSubscribeUpdateDTO
	if err = c.ShouldBindJSON(&updateDto); err != nil {
		log.Errorf("failed to bind json: %v", err)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(err))
		return
	}

	if len(updateDto.Event) <= 0 && len(updateDto.Filters) <= 0 {
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTO("no field to update"))
		return
	}

	sub.Event = updateDto.Event
	mapper := dto.Mapper{}
	var filters map[string]model.GHWebhookField
	err = mapper.Map(&filters, updateDto.Filters)
	if err != nil {
		log.Errorf("failed to bind json: %v", err)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(err))
		return
	}
	sub.Filters = filters
	if err = sub.IsValid(); err != nil {
		log.Errorf("invalid request: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}

	db = h.db.Save(&sub)
	if db.Error != nil {
		log.Errorf("failed to update webhook receiver subscribe: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}
	c.JSON(http.StatusOK, model.NewIDResponse(sub.ID))
}

func (h *GHWebhookSubscribeAPIHandler) Delete(c *gin.Context) {
	pId, err := core.UIntParam(c, "pId")
	if err != nil {
		log.Errorf("failed to convert pId: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}

	receiver := model.GHWebHookSubscribe{}
	db := h.db.First(&receiver, "id = ?", pId)
	if db.Error != nil {
		log.Errorf("failed to find webhook receiver: %v", db.Error)
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, model.NewErrorMsgDTO(http.StatusText(http.StatusBadRequest)))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}

	sub := model.GHWebHookSubscribe{}

	db = h.db.First(&sub, "id = ?", receiver.ID)
	if db.Error != nil {
		log.Errorf("failed to find webhook receiver subscribe: %v", db.Error)
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, model.NewErrorMsgDTO(http.StatusText(http.StatusNotFound)))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}

	if sub.GHWebhookReceiverID != pId {
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTO("cannot update receiver"))
		return
	}

	db = h.db.Delete(&sub)
	if db.Error != nil {
		log.Errorf("failed to delete webhook receiver subscribe: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *GHWebhookSubscribeAPIHandler) List(c *gin.Context) {
	page := core.ParsePagination(c)

	rsqlQuery := core.NewRSQLHelper()
	err := rsqlQuery.ParseFilter(GHWebhookSubscribeSearchDTO{}, c)
	if err != nil {
		log.Errorf("failed to parse filter: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}
	var subs []model.GHWebHookSubscribe
	db := h.db.Order(rsqlQuery.SortSQL).Where(rsqlQuery.FilterSQL, rsqlQuery.Arguments...).
		Limit(page.Size).Offset((page.Page - 1) * page.Size).Find(&subs)

	if db.Error != nil {
		log.Errorf("failed to find subs: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}
	var subDTOs []GHWebhookSubscribeSearchDTO
	mapper := dto.Mapper{}
	err = mapper.Map(&subDTOs, subs)
	if err != nil {
		log.Errorf("failed to map: %v", err)
		c.JSON(http.StatusInternalServerError, model.NewErrorMsgDTOFromErr(err))
		return
	}

	c.JSON(http.StatusOK, model.NewListResponse(subDTOs))
}
