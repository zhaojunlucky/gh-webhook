package core

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_ParsePagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "?page=2&size=20", nil)

	pagination := ParsePagination(c)
	if pagination.Page != 2 {
		t.Fatal("page should be equal 2")
	}
	if pagination.Size != 20 {
		t.Fatal("size should be equal 20")
	}
}

func Test_ParsePaginationDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "", nil)

	pagination := ParsePagination(c)
	if pagination.Page != 1 {
		t.Fatal("page should be equal 1")
	}
	if pagination.Size != 100 {
		t.Fatal("size should be equal 100")
	}

	c.Request, _ = http.NewRequest("GET", "?page=1&size=100", nil)

	pagination = ParsePagination(c)
	if pagination.Page != 1 {
		t.Fatal("page should be equal 1")
	}
	if pagination.Size != 100 {
		t.Fatal("size should be equal 100")
	}
}
