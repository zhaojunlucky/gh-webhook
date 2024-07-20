package core

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Pagination struct {
	Page int `form:"page"`
	Size int `form:"size"`
}

func ParsePagination(c *gin.Context) *Pagination {
	var p Pagination
	if err := c.ShouldBindQuery(&p); err != nil {
		log.Error(err)

	}

	if p.Size == 0 {
		p.Size = 100
	}
	if p.Page == 0 {
		p.Page = 1
	}

	return &p
}
