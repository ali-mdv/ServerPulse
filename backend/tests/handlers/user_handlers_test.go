package handlers_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/models"
	"server-monitoring/tests/testutil"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func newUserTestRouter(svc *stubUserService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handlers.NewUserHandler(svc)
	r.GET("/users", h.GetUsers)
	r.GET("/users/:userID", h.GetUserByID)
	r.POST("/users", h.CreateUser)
	r.PUT("/users/:userID", h.UpdateUser)
	r.GET("/me", func(c *gin.Context) { c.Set("userID", "stub-id"); c.Next() }, h.GetUserProfile)
	return r
}

func TestUserHandlers_GetUsers(t *testing.T) {
	users := []models.User{{Email: "a@b.c"}, {Email: "d@e.f"}}
	stub := &stubUserService{list: users}
	r := newUserTestRouter(stub)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("a@b.c")) || !bytes.Contains(w.Body.Bytes(), []byte("d@e.f")) {
		t.Fatalf("missing users: %s", w.Body.String())
	}
}

func TestUserHandlers_GetUsers_Error(t *testing.T) {
	stub := &stubUserService{listErr: errors.New("db down")}
	r := newUserTestRouter(stub)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/users", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code = %d, want 500", w.Code)
	}
}

func TestUserHandlers_GetUserByID_NotFound(t *testing.T) {
	stub := &stubUserService{getByIDErr: mongo.ErrNoDocuments}
	r := newUserTestRouter(stub)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/users/000000000000000000000000", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", w.Code)
	}
}

func TestUserHandlers_CreateUser_Success(t *testing.T) {
	stub := &stubUserService{createUser: nil}
	r := newUserTestRouter(stub)
	body := testutil.MustJSON(t, dtos.CreateUserDTO{Email: "new@user.io", Password: "longenough"})
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("code = %d, want 201; body = %s", w.Code, w.Body.String())
	}
}

func TestUserHandlers_CreateUser_BadJSON(t *testing.T) {
	stub := &stubUserService{}
	r := newUserTestRouter(stub)
	body := bytes.NewBufferString(`{"email":"bad","password":"x"}`) // password too short
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("code = %d, want 422", w.Code)
	}
}

func TestUserHandlers_GetProfile_RequiresUserID(t *testing.T) {
	stub := &stubUserService{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handlers.NewUserHandler(stub)
	// No middleware setting userID — the handler must refuse with 403.
	r.GET("/me", h.GetUserProfile)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/me", nil))
	if w.Code != http.StatusForbidden {
		t.Fatalf("code = %d, want 403", w.Code)
	}
}