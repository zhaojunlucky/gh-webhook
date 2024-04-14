package core

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Pagination struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

func ParsePagination(c *gin.Context) *Pagination {
	var p Pagination
	if err := c.ShouldBindQuery(&p); err != nil {
		log.Error(err)
		p.Page = 1
		p.Size = 100
	}

	return &p
}
