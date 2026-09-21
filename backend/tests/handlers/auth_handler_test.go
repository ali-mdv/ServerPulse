package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/handlers"
	apperrors "server-monitoring/pkg/errors"
	"server-monitoring/tests/testutil"

	"github.com/gin-gonic/gin"
)

func newAuthTestRouter(svc *stubAuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/login", handlers.NewAuthHandler(svc).Login)
	return r
}

func TestAuthHandler_Login_Success(t *testing.T) {
	stub := &stubAuthService{token: "tok123"}
	r := newAuthTestRouter(stub)

	body := testutil.MustJSON(t, dtos.LoginDTO{Email: "a@b.c", Password: "secret"})
	req := httptest.NewRequest("POST", "/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"token":"tok123"`) {
		t.Fatalf("body = %s", w.Body.String())
	}
	if stub.got.Email != "a@b.c" || stub.got.Password != "secret" {
		t.Fatalf("DTO not forwarded: %+v", stub.got)
	}
}

func TestAuthHandler_Login_BadJSON(t *testing.T) {
	stub := &stubAuthService{}
	r := newAuthTestRouter(stub)

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"email": "not-an-email"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("code = %d, want 422", w.Code)
	}
}

func TestAuthHandler_Login_Unauthorized(t *testing.T) {
	stub := &stubAuthService{err: apperrors.ErrUnauthorized}
	r := newAuthTestRouter(stub)

	body := testutil.MustJSON(t, dtos.LoginDTO{Email: "a@b.c", Password: "x"})
	req := httptest.NewRequest("POST", "/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", w.Code)
	}
}

func TestAuthHandler_Login_InternalError(t *testing.T) {
	stub := &stubAuthService{err: apperrors.ErrInternalServer}
	r := newAuthTestRouter(stub)

	body := testutil.MustJSON(t, dtos.LoginDTO{Email: "a@b.c", Password: "x"})
	req := httptest.NewRequest("POST", "/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code = %d, want 500", w.Code)
	}
}

// Compile-time sanity: mustJSON alias from helpers file is used by
// other handler test files; referencing json here keeps the import
// honest if someone deletes every other usage.
var _ = json.Marshal