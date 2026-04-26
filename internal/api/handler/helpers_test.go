package handler

import (
	"encoding/json"
	"goschool/pkg/httpx"
	"net/http"
	"net/http/httptest"
	"reflect"
)

func is500Error(rr *httptest.ResponseRecorder) bool {
	var payload *httpx.APIError
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		return false
	}
	if payload.Status != http.StatusInternalServerError {
		return false
	}
	if !reflect.DeepEqual(payload, httpx.ErrUnknownInternal) {
		return false
	}
	return rr.Code == http.StatusInternalServerError
}
