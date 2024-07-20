package core

import (
	"errors"
	"fmt"
	"gh-webhook/pkg/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

func UIntParam(c *gin.Context, name string) (uint, error) {
	value, ok := c.Params.Get(name)
	if !ok {
		return 0, errors.New("param not found")
	}
	var intVal uint
	n, err := fmt.Sscanf(value, "%d", &intVal)
	if err != nil || n != 1 {
		return 0, err
	}
	return intVal, nil
}

func GetModel[T any](c *gin.Context, db *gorm.DB, m *T, conds ...any) bool {
	ret := db.First(m, conds...)
	if ret.Error != nil {
		log.Errorf("failed to find webhook receiver: %v", db.Error)
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, model.NewErrorMsgDTO(http.StatusText(http.StatusNotFound)))
			return false
		}
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return false
	}
	return true
}

func GetPathVarUInt(c *gin.Context, name string) *uint {
	id, err := UIntParam(c, name)
	if err != nil {
		log.Errorf("failed to convert id: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return nil
	}
	return &id
}

func SearchModel[S any, T any](c *gin.Context, db *gorm.DB, s S, results *[]T) bool {
	page := ParsePagination(c)

	rsqlQuery := NewRSQLHelper()
	err := rsqlQuery.ParseFilter(s, c)
	if err != nil {
		log.Errorf("failed to parse filter: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return false
	}

	ret := db.Order(rsqlQuery.SortSQL).Where(rsqlQuery.FilterSQL, rsqlQuery.Arguments...).
		Limit(page.Size).Offset((page.Page - 1) * page.Size).Find(results)

	if ret.Error != nil {
		log.Errorf("failed to find receivers: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return false
	}
	return true
}
