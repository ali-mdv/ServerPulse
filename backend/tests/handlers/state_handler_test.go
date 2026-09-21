package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"server-monitoring/internal/handlers"
	"server-monitoring/internal/models"
	apperrors "server-monitoring/pkg/errors"

	"github.com/gin-gonic/gin"
)

func newStateTestRouter(svc *stubStateService2, srv *minimalStateServerSvc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handlers.NewStateHandler(svc, srv)
	r.GET("/state/services", h.GetServicesState)
	r.GET("/state/system", h.GetSystemState)
	return r
}

func TestStateHandler_GetServicesState_NilProviders(t *testing.T) {
	srv := &minimalStateServerSvc{}
	st := &stubStateService2{} // both pm2/docker nil
	r := newStateTestRouter(st, srv)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/state/services", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d", w.Code)
	}
	var body struct {
		Pm2    models.ProviderState `json:"pm2"`
		Docker models.ProviderState `json:"docker"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Pm2.UpdatedAt.IsZero() {
		t.Fatalf("expected zero provider when nil in repo: %+v", body.Pm2)
	}
}

func TestStateHandler_GetServicesState_Populated(t *testing.T) {
	srv := &minimalStateServerSvc{}
	st := &stubStateService2{
		pm2:    &models.ProviderState{ServerID: "local", Provider: "pm2", Available: true, UpdatedAt: time.Now().UTC()},
		docker: &models.ProviderState{ServerID: "local", Provider: "docker", Available: false, UpdatedAt: time.Now().UTC()},
	}
	r := newStateTestRouter(st, srv)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/state/services", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d", w.Code)
	}
	if !contains(w.Body.String(), `"available":true`) {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestStateHandler_GetServicesState_Error(t *testing.T) {
	srv := &minimalStateServerSvc{}
	st := &stubStateService2{stateErr: apperrors.ErrInternalServer}
	r := newStateTestRouter(st, srv)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/state/services", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code = %d, want 500", w.Code)
	}
}

func TestStateHandler_GetSystemState_NilUsage(t *testing.T) {
	srv := &minimalStateServerSvc{}
	st := &stubStateService2{}
	r := newStateTestRouter(st, srv)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/state/system", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d", w.Code)
	}
	if !contains(w.Body.String(), `"systemUsage":null`) {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestStateHandler_GetSystemState_Populated(t *testing.T) {
	srv := &minimalStateServerSvc{}
	st := &stubStateService2{usage: &models.ServerUsageState{
		ServerID: "local", UpdatedAt: time.Now().UTC(),
		Usage: models.SystemUsage{CpuUsage: 7},
	}}
	r := newStateTestRouter(st, srv)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/state/system", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d", w.Code)
	}
	if !contains(w.Body.String(), `"cpuUsage":7`) {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestStateHandler_GetSystemState_Error(t *testing.T) {
	srv := &minimalStateServerSvc{}
	st := &stubStateService2{usageErr: apperrors.ErrInternalServer}
	r := newStateTestRouter(st, srv)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/state/system", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code = %d, want 500", w.Code)
	}
}