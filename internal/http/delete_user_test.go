package http

import (
	stdhttp "net/http"
	"testing"
)

const validDeleteCredentials = `{"login":"ana","password":"password123"}`

func requireDeleteResponse(t *testing.T, resCode, wantStatus int, body map[string]any, wantMessage string) {
	t.Helper()

	if resCode != wantStatus {
		t.Fatalf("expected status %d, got %d", wantStatus, resCode)
	}
	if got, want := body["code"], float64(wantStatus); got != want {
		t.Fatalf("expected response code %v, got %v", want, got)
	}
	if got, want := body["message"], wantMessage; got != want {
		t.Fatalf("expected message %q, got %q", want, got)
	}
}

func requireDeleteError(t *testing.T, body map[string]any, field, message string) {
	t.Helper()

	errors, ok := body["errors"].(map[string]any)
	if !ok {
		t.Fatalf("expected errors object, got %#v", body["errors"])
	}
	if got, want := errors[field], message; got != want {
		t.Fatalf("expected %s error %q, got %q", field, want, got)
	}
}

func TestDeleteUserReturnsSuccessAndRemovesUser(t *testing.T) {
	router := newTestRouter(t)
	createTestUser(t, router)

	res := performDelete(t, router, "1", validDeleteCredentials)
	body := decodeBody(t, res)
	requireDeleteResponse(t, res.Code, stdhttp.StatusOK, body, "user deleted successfully")

	res = performDelete(t, router, "1", validDeleteCredentials)
	body = decodeBody(t, res)
	requireDeleteResponse(t, res.Code, stdhttp.StatusNotFound, body, "user not found")
	requireDeleteError(t, body, "not_found", "user not found")
}

func TestDeleteUserRejectsInvalidCredentialsAndPreservesUser(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		field   string
		message string
	}{
		{name: "wrong login", payload: `{"login":"other","password":"password123"}`, field: "login", message: "wrong login"},
		{name: "wrong password", payload: `{"login":"ana","password":"another-password"}`, field: "password", message: "wrong password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newTestRouter(t)
			createTestUser(t, router)

			res := performDelete(t, router, "1", tt.payload)
			body := decodeBody(t, res)
			requireDeleteResponse(t, res.Code, stdhttp.StatusUnauthorized, body, tt.message)
			requireDeleteError(t, body, tt.field, tt.message)

			res = performDelete(t, router, "1", validDeleteCredentials)
			body = decodeBody(t, res)
			requireDeleteResponse(t, res.Code, stdhttp.StatusOK, body, "user deleted successfully")
		})
	}
}

func TestDeleteUserValidatesRequest(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		payload string
		message string
		field   string
		error   string
	}{
		{name: "invalid JSON", id: "1", payload: `{"login":`, message: "invalid json", field: "json", error: "invalid"},
		{name: "missing login", id: "1", payload: `{"password":"password123"}`, message: "validation failed", field: "login", error: "login cannot be empty"},
		{name: "blank login", id: "1", payload: `{"login":"  ","password":"password123"}`, message: "validation failed", field: "login", error: "login cannot be empty"},
		{name: "missing password", id: "1", payload: `{"login":"ana"}`, message: "validation failed", field: "password", error: "password cannot be empty"},
		{name: "blank password", id: "1", payload: `{"login":"ana","password":"  "}`, message: "validation failed", field: "password", error: "password cannot be empty"},
		{name: "non-numeric ID", id: "abc", payload: validDeleteCredentials, message: "invalid user id", field: "invalid", error: "invalid user id"},
		{name: "zero ID", id: "0", payload: validDeleteCredentials, message: "invalid user id", field: "invalid", error: "invalid user id"},
		{name: "negative ID", id: "-1", payload: validDeleteCredentials, message: "invalid user id", field: "invalid", error: "invalid user id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := performDelete(t, newTestRouter(t), tt.id, tt.payload)
			body := decodeBody(t, res)
			requireDeleteResponse(t, res.Code, stdhttp.StatusBadRequest, body, tt.message)
			requireDeleteError(t, body, tt.field, tt.error)
		})
	}
}

func TestDeleteUserReturnsNotFoundForMissingUser(t *testing.T) {
	res := performDelete(t, newTestRouter(t), "99", validDeleteCredentials)
	body := decodeBody(t, res)
	requireDeleteResponse(t, res.Code, stdhttp.StatusNotFound, body, "user not found")
	requireDeleteError(t, body, "not_found", "user not found")
}

func TestDeleteUserWithoutIDDoesNotMatchRoute(t *testing.T) {
	res := performDelete(t, newTestRouter(t), "", validDeleteCredentials)
	if res.Code != stdhttp.StatusNotFound {
		t.Fatalf("expected unmatched route status %d, got %d: %s", stdhttp.StatusNotFound, res.Code, res.Body.String())
	}
}
