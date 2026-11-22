package http

import (
	stdhttp "net/http"
	"testing"
)

func TestUpdateUserPartialAndPasswordHash(t *testing.T) {
	router := newTestRouter(t)
	created := performPost(t, router, `{"name":"Ana","login":"ana","email":"ana@example.com","password":"password123"}`)
	if created.Code != stdhttp.StatusCreated {
		t.Fatalf("create user: %d: %s", created.Code, created.Body.String())
	}

	res := performPatch(t, router, "1", `{"name":" Ana Maria ","password":"newpassword123"}`)
	if res.Code != stdhttp.StatusNoContent {
		t.Fatalf("expected status %d, got %d: %s", stdhttp.StatusNoContent, res.Code, res.Body.String())
	}
	if res.Body.Len() != 0 {
		t.Fatalf("expected empty response body, got %q", res.Body.String())
	}

	res = performDelete(t, router, "1", `{"login":"ana","password":"newpassword123"}`)
	if res.Code != stdhttp.StatusOK {
		t.Fatalf("updated password was not verifiable: %d: %s", res.Code, res.Body.String())
	}
}

func TestUpdateUserValidationAndNotFound(t *testing.T) {
	router := newTestRouter(t)

	for _, tt := range []struct {
		name string
		id   string
		body string
	}{
		{name: "invalid id", id: "invalid", body: `{"name":"Ana"}`},
		{name: "invalid json", id: "1", body: `{"name":`},
		{name: "empty body", id: "1", body: `{}`},
		{name: "blank login", id: "1", body: `{"login":"  "}`},
		{name: "invalid email", id: "1", body: `{"email":"invalid"}`},
		{name: "short password", id: "1", body: `{"password":"short"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			res := performPatch(t, router, tt.id, tt.body)
			if res.Code != stdhttp.StatusBadRequest {
				t.Fatalf("expected status %d, got %d: %s", stdhttp.StatusBadRequest, res.Code, res.Body.String())
			}
		})
	}

	res := performPatch(t, router, "99", `{"name":"Ana"}`)
	if res.Code != stdhttp.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", stdhttp.StatusNotFound, res.Code, res.Body.String())
	}
}

func TestUpdateUserRejectsDuplicateLogin(t *testing.T) {
	router := newTestRouter(t)
	for _, payload := range []string{
		`{"name":"Ana","login":"ana","password":"password123"}`,
		`{"name":"Beto","login":"beto","password":"password123"}`,
	} {
		res := performPost(t, router, payload)
		if res.Code != stdhttp.StatusCreated {
			t.Fatalf("create user: %d: %s", res.Code, res.Body.String())
		}
	}

	res := performPatch(t, router, "2", `{"login":"ana"}`)
	if res.Code != stdhttp.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", stdhttp.StatusConflict, res.Code, res.Body.String())
	}
}
