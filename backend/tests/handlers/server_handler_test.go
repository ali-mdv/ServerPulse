package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/handlers"
	"server-monitoring/internal/models"
	apperrors "server-monitoring/pkg/errors"
	"server-monitoring/tests/testutil"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func newServerTestRouter(srv *stubServerSvc, st *stubStateSvc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handlers.NewServerHandler(srv, st)
	r.GET("/servers", h.List)
	r.GET("/servers/:serverId", h.Get)
	r.POST("/servers", h.Create)
	r.PUT("/servers/:serverId", h.Update)
	r.DELETE("/servers/:serverId", h.Delete)
	r.POST("/servers/:serverId/api-key", h.GenerateAgentToken)
	return r
}

func TestServerHandler_List(t *testing.T) {
	id := bson.NewObjectID()
	srv := &stubServerSvc{list: []models.Server{{
		ID: id, Name: "web-01", Host: "10.0.0.1", Status: models.ServerStatusOnline,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}}}
	st := &stubStateSvc{}
	r := newServerTestRouter(srv, st)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/servers", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("web-01")) {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestServerHandler_Get_BadID(t *testing.T) {
	r := newServerTestRouter(&stubServerSvc{}, &stubStateSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/servers/not-an-objectid", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", w.Code)
	}
}

func TestServerHandler_Get_NotFound(t *testing.T) {
	srv := &stubServerSvc{getByIDErr: apperrors.ErrNotFound}
	r := newServerTestRouter(srv, &stubStateSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/servers/"+bson.NewObjectID().Hex(), nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", w.Code)
	}
}

func TestServerHandler_Get_Success(t *testing.T) {
	id := bson.NewObjectID()
	now := time.Now().UTC()
	srv := &stubServerSvc{getByID: &models.Server{
		ID: id, Name: "n", Host: "h", Status: models.ServerStatusOnline,
		CreatedAt: now, UpdatedAt: now,
	}}
	st := &stubStateSvc{usage: &models.ServerUsageState{
		ServerID:  id.Hex(),
		UpdatedAt: now,
		Usage:     models.SystemUsage{CpuUsage: 1.5},
	}}
	r := newServerTestRouter(srv, st)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/servers/"+id.Hex(), nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"usage"`)) {
		t.Fatalf("expected usage in body, got %s", w.Body.String())
	}
}

func TestServerHandler_Create(t *testing.T) {
	id := bson.NewObjectID()
	now := time.Now().UTC()
	srv := &stubServerSvc{created: &models.Server{
		ID: id, Name: "new", Host: "h", CreatedAt: now, UpdatedAt: now,
	}}
	r := newServerTestRouter(srv, &stubStateSvc{})
	w := httptest.NewRecorder()
	body := testutil.MustJSON(t, dtos.CreateServerDTO{Name: "new"})
	req := httptest.NewRequest("POST", "/servers", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("code = %d, want 201; body = %s", w.Code, w.Body.String())
	}
}

func TestServerHandler_Create_BadJSON(t *testing.T) {
	r := newServerTestRouter(&stubServerSvc{}, &stubStateSvc{})
	body := bytes.NewBufferString(`{}`) // missing required name
	req := httptest.NewRequest("POST", "/servers", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("code = %d, want 422", w.Code)
	}
}

func TestServerHandler_GenerateAgentToken(t *testing.T) {
	id := bson.NewObjectID()
	now := time.Now().UTC()
	srv := &stubServerSvc{
		updated:  &models.Server{ID: id, Name: "n", Host: "h", CreatedAt: now, UpdatedAt: now},
		genToken: "the-only-token",
	}
	r := newServerTestRouter(srv, &stubStateSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/servers/"+id.Hex()+"/api-key", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("the-only-token")) {
		t.Fatalf("token missing in body: %s", w.Body.String())
	}
}

func TestServerHandler_Delete_Forbidden(t *testing.T) {
	srv := &stubServerSvc{deleted: apperrors.ErrForbidden}
	r := newServerTestRouter(srv, &stubStateSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("DELETE", "/servers/"+bson.NewObjectID().Hex(), nil))
	if w.Code != http.StatusForbidden {
		t.Fatalf("code = %d, want 403", w.Code)
	}
}