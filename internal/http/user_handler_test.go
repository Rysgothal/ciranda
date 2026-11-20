package http

import (
	"bytes"
	"ciranda/internal/database"
	"database/sql"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	if err = database.Migrate(t.Context(), db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	gin.SetMode(gin.TestMode)

	server := NewServer(db)
	router := gin.New()
	router.POST("/users", server.CreateUser)
	router.DELETE("/users/:id", server.DeleteUser)

	return router
}

func performDelete(t *testing.T, router stdhttp.Handler, id string, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(stdhttp.MethodDelete, "/users/"+id, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	return res
}

func performPost(t *testing.T, router stdhttp.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(stdhttp.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	return res
}

func decodeBody(t *testing.T, res *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	return body
}

func TestCreateUserReturnsCreatedUser(t *testing.T) {
	router := newTestRouter(t)

	res := performPost(t, router, `{
		"name": " Ana ",
		"login": " anabanana ",
		"email": "ana@banana.com",
		"password": "AnaBanana123"
	}`)

	if res.Code != stdhttp.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", stdhttp.StatusCreated, res.Code, res.Body.String())
	}

	if got, want := res.Header().Get("Location"), "/users/1"; got != want {
		t.Fatalf("expected Location %q, got %q", want, got)
	}

	body := decodeBody(t, res)
	user, ok := body["user"].(map[string]any)
	if !ok {
		t.Fatalf("expected user object, got %#v", body["user"])
	}

	if got, want := user["id"], float64(1); got != want {
		t.Fatalf("expected id %v, got %v", want, got)
	}
	if got, want := user["name"], "Ana"; got != want {
		t.Fatalf("expected name %q, got %q", want, got)
	}
	if got, want := user["login"], "anabanana"; got != want {
		t.Fatalf("expected login %q, got %q", want, got)
	}
	if got, want := user["email"], "ana@banana.com"; got != want {
		t.Fatalf("expected email %q, got %q", want, got)
	}
	if _, exists := user["password"]; exists {
		t.Fatal("response must not expose password")
	}
}

func TestCreateUserRejectsInvalidJSON(t *testing.T) {
	res := performPost(t, newTestRouter(t), `{"name":`)

	if res.Code != stdhttp.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", stdhttp.StatusBadRequest, res.Code, res.Body.String())
	}

	body := decodeBody(t, res)
	if got, want := body["message"], "invalid json"; got != want {
		t.Fatalf("expected message %q, got %q", want, got)
	}
}

func TestCreateUserValidatesFields(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		field   string
		message string
	}{
		{
			name:    "blank name",
			payload: `{"name":"   ","login":"ana","password":"password123"}`,
			field:   "name",
			message: "name cannot be empty",
		},
		{
			name:    "blank login",
			payload: `{"name":"Ana","login":"   ","password":"password123"}`,
			field:   "login",
			message: "login cannot be empty",
		},
		{
			name:    "blank password",
			payload: `{"name":"Ana","login":"ana","password":"        "}`,
			field:   "password",
			message: "password cannot be empty",
		},
		{
			name:    "short password",
			payload: `{"name":"Ana","login":"ana","password":"short"}`,
			field:   "password",
			message: "password must be at least 8 characters",
		},
		{
			name:    "invalid email",
			payload: `{"name":"Ana","login":"ana","email":"invalid","password":"password123"}`,
			field:   "email",
			message: "invalid email",
		},
		{
			name:    "name maximum",
			payload: `{"name":"` + strings.Repeat("a", 101) + `","login":"ana","password":"password123"}`,
			field:   "name",
			message: "name must be at most 100 characters",
		},
		{
			name:    "login maximum",
			payload: `{"name":"Ana","login":"` + strings.Repeat("a", 51) + `","password":"password123"}`,
			field:   "login",
			message: "login must be at most 50 characters",
		},
		{
			name:    "email maximum",
			payload: `{"name":"Ana","login":"ana","email":"` + strings.Repeat("a", 246) + `@test.com","password":"password123"}`,
			field:   "email",
			message: "email must be at most 254 characters",
		},
		{
			name:    "password maximum",
			payload: `{"name":"Ana","login":"ana","password":"` + strings.Repeat("a", 129) + `"}`,
			field:   "password",
			message: "password must be at most 128 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := performPost(t, newTestRouter(t), tt.payload)
			if res.Code != stdhttp.StatusBadRequest {
				t.Fatalf("expected status %d, got %d: %s", stdhttp.StatusBadRequest, res.Code, res.Body.String())
			}

			body := decodeBody(t, res)
			errors, ok := body["errors"].(map[string]any)
			if !ok {
				t.Fatalf("expected errors object, got %#v", body["errors"])
			}
			if got := errors[tt.field]; got != tt.message {
				t.Errorf("expected %s error %q, got %q", tt.field, tt.message, got)
			}
		})
	}
}

func TestCreateUserRejectsDuplicateLogin(t *testing.T) {
	router := newTestRouter(t)
	payload := `{"name":"Ana","login":"ana","password":"password123"}`

	first := performPost(t, router, payload)
	if first.Code != stdhttp.StatusCreated {
		t.Fatalf("first request failed: %d: %s", first.Code, first.Body.String())
	}

	second := performPost(t, router, payload)
	if second.Code != stdhttp.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", stdhttp.StatusConflict, second.Code, second.Body.String())
	}

	body := decodeBody(t, second)
	errors, ok := body["errors"].(map[string]any)
	if !ok {
		t.Fatalf("expected errors object, got %#v", body["errors"])
	}
	if got, want := errors["login"], "user with this login already exists"; got != want {
		t.Fatalf("expected conflict message %q, got %q", want, got)
	}
}

func TestDeleteUserWrongLogin(t *testing.T) {
	router := newTestRouter(t)
	_ = performPost(t, router, `{
		"name": " Ana ",
		"login": " anabanana ",
		"email": "ana@banana.com",
		"password": "AnaBanana123"
	}`)

	res := performDelete(t, router, "1", `{
		"login": " beto betando ",
		"password": "AnaBanana123"
	}`)

	if res.Code != stdhttp.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d: %s", stdhttp.StatusUnauthorized, res.Code, res.Body.String())
	}
}

func TestDeleteUserWrongPassword(t *testing.T) {
	router := newTestRouter(t)

	_ = performPost(t, router, `{
		"name": " Ana ",
		"login": " anabanana ",
		"email": "ana@banana.com",
		"password": "AnaBanana123"
	}`)

	res := performDelete(t, router, "1", `{
		"login": " anabanana ",
		"password": "pass123456"
	}`)

	if res.Code != stdhttp.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d: %s", stdhttp.StatusUnauthorized, res.Code, res.Body.String())
	}
}

func TestDeleteUserNoID(t *testing.T) {
	router := newTestRouter(t)

	res := performDelete(t, router, "", `{
		"login": " anabanana ",
		"password": "AnaBanana123"
	}`)

	if res.Code != stdhttp.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", stdhttp.StatusNotFound, res.Code, res.Body.String())
	}
}

func TestDeleteUserOK(t *testing.T) {
	router := newTestRouter(t)

	resPost := performPost(t, router, `{
		"name": " Ana ",
		"login": " anabanana ",
		"email": "ana@banana.com",
		"password": "AnaBanana123"
	}`)

	if got, want := resPost.Header().Get("Location"), "/users/1"; got != want {
		t.Fatalf("expected Location %q, got %q", want, got)
	}

	res := performDelete(t, router, "1", `{
		"login": " anabanana ",
		"password": "AnaBanana123"
	}`)

	if res.Code != stdhttp.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", stdhttp.StatusOK, res.Code, res.Body.String())
	}

	res2 := performDelete(t, router, "1", `{
		"login": " anabanana ",
		"password": "AnaBanana123"
	}`)

	if res2.Code != stdhttp.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", stdhttp.StatusNotFound, res2.Code, res2.Body.String())
	}
}
