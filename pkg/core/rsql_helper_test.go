package core

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"
)

type TestRSQLDTO struct {
	ID        uint      `json:"id" rsql:"id,filter,sort"`
	CreatedAt time.Time `json:"createdAt" `
	UpdatedAt time.Time `json:"updatedAt" `
	Web       string    `json:"web" rsql:"web,filter,sort"`
	API       string    `json:"api" rsql:"api,filter,sort"`
	Name      string    `json:"name" rsql:"name,filter,sort"`
}

func Test_ParseFilterInvalidField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "?filter=name==\"jun\";age==37", nil)

	helper := NewRSQLHelper()

	err := helper.ParseFilter(TestRSQLDTO{}, c)
	if err == nil {
		t.Fatal("should be failed")
	}

	if err.Error() != "field age is not allowed to filter" {
		t.Fatal(err)
	}

}

func TestRSQLHelper_ParseFilterAnd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "?filter=name==\"jun\";id==1;web==\"github.com\"", nil)

	helper := NewRSQLHelper()

	err := helper.ParseFilter(TestRSQLDTO{}, c)
	if err != nil {
		t.Fatal(err)
	}
	expectedSQL := "name = ? and id = ? and web = ?"
	if helper.FilterSQL != expectedSQL {
		t.Fatalf("expected sql (%s) != actual sql (%s)", expectedSQL, helper.FilterSQL)
	}

	expectedArgs := []interface{}{"jun", uint64(1), "github.com"}
	if !slices.Equal(helper.Arguments, expectedArgs) {
		t.Fatalf("expected args (%v) != actual args (%v)", expectedArgs, helper.Arguments)
	}

	if len(helper.SortSQL) != 0 {
		t.Fatal("sort sql should be null")
	}
}

func TestRSQLHelper_ParseFilterOr(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "?filter=name==\"jun\";id==1,web==\"github.com\"", nil)

	helper := NewRSQLHelper()

	err := helper.ParseFilter(TestRSQLDTO{}, c)
	if err != nil {
		t.Fatal(err)
	}
	expectedSQL := "(name = ? and id = ?) or web = ?"
	if helper.FilterSQL != expectedSQL {
		t.Fatalf("expected sql (%s) != actual sql (%s)", expectedSQL, helper.FilterSQL)
	}

	expectedArgs := []interface{}{"jun", uint64(1), "github.com"}
	if !slices.Equal(helper.Arguments, expectedArgs) {
		t.Fatalf("expected args (%v) != actual args (%v)", expectedArgs, helper.Arguments)
	}

	if len(helper.SortSQL) != 0 {
		t.Fatal("sort sql should be null")
	}
}

func TestRSQLHelper_ParseSort(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "?sort=name,asc;id,desc", nil)

	helper := NewRSQLHelper()

	err := helper.ParseFilter(TestRSQLDTO{}, c)
	if err != nil {
		t.Fatal(err)
	}

	if helper.SortSQL != "name asc,id desc" {
		t.Fatalf("expected sql (%s) != actual sql (%s)", "name asc, id desc", helper.SortSQL)
	}
}
