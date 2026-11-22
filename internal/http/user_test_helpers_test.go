package http

import (
	"bytes"
	"ciranda/internal/database"
	"database/sql"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
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
	router.PATCH("/users/:id", server.UpdateUser)

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

func performPatch(t *testing.T, router stdhttp.Handler, id string, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(stdhttp.MethodPatch, "/users/"+id, bytes.NewBufferString(body))
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

func createTestUser(t *testing.T, router stdhttp.Handler) {
	t.Helper()

	res := performPost(t, router, `{"name":"Ana","login":"ana","email":"ana@example.com","password":"password123"}`)
	if res.Code != stdhttp.StatusCreated {
		t.Fatalf("create test user: expected status %d, got %d: %s", stdhttp.StatusCreated, res.Code, res.Body.String())
	}
}
