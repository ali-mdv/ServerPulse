package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"server-monitoring/internal/handlers"
	"server-monitoring/internal/models"
	apperrors "server-monitoring/pkg/errors"

	"github.com/gin-gonic/gin"
)

func newPM2TestRouter(svc *stubPM2Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handlers.NewPM2Handler(svc)
	r.GET("/pm2/services", h.ProcessList)
	r.GET("/pm2/services/:id", h.ProcessDetail)
	r.POST("/pm2/services/:id/start", h.StartProcess)
	r.POST("/pm2/services/:id/stop", h.StopProcess)
	r.POST("/pm2/services/:id/restart", h.RestartProcess)
	r.GET("/pm2/services/:id/logs", h.GetProcessLogs)
	return r
}

func TestPM2Handler_List_Success(t *testing.T) {
	svc := &stubPM2Service{list: []models.PM2Process{{Name: "api", PMID: 0}}}
	r := newPM2TestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/pm2/services", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d", w.Code)
	}
	if !contains(w.Body.String(), `"available":true`) {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestPM2Handler_List_UnavailableDegrades(t *testing.T) {
	svc := &stubPM2Service{listErr: apperrors.New(503, "pm2 daemon unreachable")}
	r := newPM2TestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/pm2/services", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", w.Code)
	}
	if !contains(w.Body.String(), `"available":false`) {
		t.Fatalf("body = %s", w.Body.String())
	}
	if !contains(w.Body.String(), `"processes":[]`) {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestPM2Handler_List_AppError(t *testing.T) {
	svc := &stubPM2Service{listErr: apperrors.ErrNotFound}
	r := newPM2TestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/pm2/services", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", w.Code)
	}
}

func TestPM2Handler_ProcessDetail_BadID(t *testing.T) {
	svc := &stubPM2Service{}
	r := newPM2TestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/pm2/services/notanumber", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", w.Code)
	}
}

func TestPM2Handler_ProcessDetail_NotFound(t *testing.T) {
	svc := &stubPM2Service{procErr: apperrors.ErrNotFound}
	r := newPM2TestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/pm2/services/3", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", w.Code)
	}
}

func TestPM2Handler_ProcessDetail_Success(t *testing.T) {
	svc := &stubPM2Service{proc: &models.PM2Process{Name: "api", PMID: 3}}
	r := newPM2TestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/pm2/services/3", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d", w.Code)
	}
}

func TestPM2Handler_Start_Success(t *testing.T) {
	svc := &stubPM2Service{}
	r := newPM2TestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/pm2/services/1/start", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", w.Code)
	}
}

func TestPM2Handler_Stop_BadID(t *testing.T) {
	svc := &stubPM2Service{}
	r := newPM2TestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/pm2/services/abc/stop", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", w.Code)
	}
}

func TestPM2Handler_Restart_NotFound(t *testing.T) {
	svc := &stubPM2Service{restartErr: apperrors.ErrNotFound}
	r := newPM2TestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/pm2/services/1/restart", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", w.Code)
	}
}

func TestPM2Handler_Logs_BadID(t *testing.T) {
	svc := &stubPM2Service{}
	r := newPM2TestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/pm2/services/x/logs", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", w.Code)
	}
}

func TestPM2Handler_Logs_Success(t *testing.T) {
	svc := &stubPM2Service{logs: "line1\nline2"}
	r := newPM2TestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/pm2/services/1/logs", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", w.Code)
	}
	if !contains(w.Body.String(), "line1") {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}