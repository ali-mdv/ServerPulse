package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"server-monitoring/internal/handlers"
	"server-monitoring/internal/models"
	apperrors "server-monitoring/pkg/errors"

	"github.com/gin-gonic/gin"
)

func newHistoryTestRouter(svc *stubHistoryService, srv *minimalHistoryServerSvc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handlers.NewHistoryHandler(svc, srv)
	r.GET("/history/:provider/services", h.ListServices)
	r.GET("/history/:provider/services/:serviceId", h.GetSeries)
	return r
}

func TestHistoryHandler_ListServices_InvalidProvider(t *testing.T) {
	r := newHistoryTestRouter(&stubHistoryService{}, &minimalHistoryServerSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/history/unknown/services", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", w.Code)
	}
}

func TestHistoryHandler_ListServices_Success(t *testing.T) {
	r := newHistoryTestRouter(&stubHistoryService{
		tracked: []modelsSnapshotMeta{{ServerID: "local", Provider: "pm2", ServiceID: "0", Name: "api"}},
	}, &minimalHistoryServerSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/history/pm2/services", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	var body struct {
		Provider string                 `json:"provider"`
		Services []modelsSnapshotMeta   `json:"services"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Provider != "pm2" || len(body.Services) != 1 {
		t.Fatalf("bad body: %+v", body)
	}
}

func TestHistoryHandler_GetSeries_BadProvider(t *testing.T) {
	r := newHistoryTestRouter(&stubHistoryService{}, &minimalHistoryServerSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/history/unknown/services/0", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", w.Code)
	}
}

func TestHistoryHandler_GetSeries_MissingServiceID(t *testing.T) {
	r := newHistoryTestRouter(&stubHistoryService{}, &minimalHistoryServerSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/history/pm2/services/%20", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", w.Code)
	}
}

func TestHistoryHandler_GetSeries_InvalidRange(t *testing.T) {
	r := newHistoryTestRouter(&stubHistoryService{}, &minimalHistoryServerSvc{})
	q := url.Values{}
	q.Set("from", "not-a-time")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/history/pm2/services/0?"+q.Encode(), nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", w.Code)
	}
}

func TestHistoryHandler_GetSeries_RangeTooLarge(t *testing.T) {
	r := newHistoryTestRouter(&stubHistoryService{}, &minimalHistoryServerSvc{})
	now := time.Now().UTC().Format(time.RFC3339)
	q := url.Values{}
	q.Set("from", time.Now().Add(-30*24*time.Hour).UTC().Format(time.RFC3339))
	q.Set("to", now)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/history/pm2/services/0?"+q.Encode(), nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", w.Code)
	}
}

func TestHistoryHandler_GetSeries_BadBucket(t *testing.T) {
	r := newHistoryTestRouter(&stubHistoryService{}, &minimalHistoryServerSvc{})
	q := url.Values{}
	q.Set("bucket", "notanumber")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/history/pm2/services/0?"+q.Encode(), nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", w.Code)
	}
}

func TestHistoryHandler_GetSeries_Success(t *testing.T) {
	r := newHistoryTestRouter(&stubHistoryService{
		points: []modelsServiceSnapshot{{Meta: modelsSnapshotMeta{ServiceID: "0", Provider: "pm2"}, CPU: 1.0}},
	}, &minimalHistoryServerSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/history/pm2/services/0", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200; body = %s", w.Code, w.Body.String())
	}
}

func TestHistoryHandler_GetSeries_ServiceError(t *testing.T) {
	r := newHistoryTestRouter(&stubHistoryService{pointsErr: apperrors.ErrInternalServer}, &minimalHistoryServerSvc{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/history/pm2/services/0", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code = %d, want 500", w.Code)
	}
}

// aliases keep the history handler tests free of an explicit
// import of models just for shape literals.
type (
	modelsSnapshotMeta   = models.SnapshotMeta
	modelsServiceSnapshot = models.ServiceSnapshot
)