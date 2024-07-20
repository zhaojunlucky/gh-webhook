package model

import (
	"errors"
	"testing"
)

func Test_NewIDResponse(t *testing.T) {
	idResp := NewIDResponse(100)
	if idResp.ID != 100 {
		t.Fatal("id should be equal 100")
	}
}

func Test_NewErrMsgsFromErr(t *testing.T) {

	errs := []error{errors.New("err1"), errors.New("err2")}
	errMsgs := NewErrMsgsFromErr(errs...)
	if len(errMsgs) != 2 {
		t.Fatal("should be 2 error messages")
	}

	if errMsgs[0].Message != "err1" {
		t.Fatal("message should be err1")
	}

	if errMsgs[1].Message != "err2" {
		t.Fatal("message should be err2")
	}
}

func Test_NewErrMsgs(t *testing.T) {
	errs := []string{"err1", "err2"}
	errMsgs := NewErrMsgs(errs...)
	if len(errMsgs) != 2 {
		t.Fatal("should be 2 error messages")
	}

	if errMsgs[0].Message != "err1" {
		t.Fatal("message should be err1")
	}

	if errMsgs[1].Message != "err2" {
		t.Fatal("message should be err2")
	}
}

func Test_NewErrorMsgDTO(t *testing.T) {
	errMsg := NewErrorMsgDTO("err1")
	if errMsg.ErrorMessages[0].Message != "err1" {
		t.Fatal("message should be err1")
	}
}

func Test_NewErrorMsgDTOFromErr(t *testing.T) {
	errs := []error{errors.New("err1"), errors.New("err2")}
	errMsg := NewErrorMsgDTOFromErr(errs...)
	if errMsg.ErrorMessages[0].Message != "err1" {
		t.Fatal("message should be err1")
	}
	if errMsg.ErrorMessages[1].Message != "err2" {
		t.Fatal("message should be err2")
	}
}

func Test_NewListResponse(t *testing.T) {
	errMsgs := []ErrorMessage{
		{Message: "err1"},
		{Message: "err2"},
	}

	listResp := NewListResponse(errMsgs)

	if listResp.EntryCount != 2 {
		t.Fatal("entry count should be 2")
	}

	if len(listResp.Entries) != 2 {
		t.Fatal("entries should be 2")
	}

	if listResp.Entries[0].Message != "err1" {
		t.Fatal("message should be err1")
	}

	if listResp.Entries[1].Message != "err2" {
		t.Fatal("message should be err2")
	}
}
