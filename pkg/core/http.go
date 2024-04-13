package core

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
)

func UIntParam(c *gin.Context, name string) (uint, error) {
	value, ok := c.Params.Get(name)
	if !ok {
		return 0, errors.New("param not found")
	}
	var intVal uint
	n, err := fmt.Sscanf(value, "%u", &intVal)
	if err != nil || n != 1 {
		return 0, errors.New("invalid param")
	}
	return intVal, nil
}
