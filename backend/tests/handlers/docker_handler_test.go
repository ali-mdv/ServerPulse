package handlers_test

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"server-monitoring/internal/handlers"
	"server-monitoring/internal/models"
	apperrors "server-monitoring/pkg/errors"

	"github.com/gin-gonic/gin"
)

func newDockerTestRouter(svc *stubDockerService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handlers.NewDockerHandler(svc)
	r.GET("/containers", h.GetDockerContainers)
	r.POST("/containers/:containerId/start", h.StartContainer)
	return r
}

func TestDockerHandler_List_Success(t *testing.T) {
	svc := &stubDockerService{containers: []models.DockerContainer{
		{ID: "c1", Name: "web"},
	}}
	r := newDockerTestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/containers", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", w.Code)
	}
	var body struct {
		Containers []models.DockerContainer `json:"containers"`
		Available  bool                    `json:"available"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Available {
		t.Fatal("expected available=true")
	}
	if len(body.Containers) != 1 || body.Containers[0].ID != "c1" {
		t.Fatalf("bad body: %+v", body)
	}
}

func TestDockerHandler_List_DockerUnavailable_Degrades(t *testing.T) {
	// The handler matches strings.Contains "Cannot connect to the Docker daemon".
	svc := &stubDockerService{containersErr: errors.New("Cannot connect to the Docker daemon at unix:///var/run/docker.sock")}
	r := newDockerTestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/containers", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200 (degraded, not 503)", w.Code)
	}
	var body struct {
		Containers []any `json:"containers"`
		Available  bool `json:"available"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Available {
		t.Fatal("expected available=false")
	}
	if len(body.Containers) != 0 {
		t.Fatalf("expected empty list, got %+v", body.Containers)
	}
}

func TestDockerHandler_List_NetOpError_Degrades(t *testing.T) {
	svc := &stubDockerService{
		containersErr: &net.OpError{Op: "dial", Net: "unix", Err: errors.New("connection refused")},
	}
	r := newDockerTestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/containers", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200 (degraded)", w.Code)
	}
}

func TestDockerHandler_List_AppError(t *testing.T) {
	svc := &stubDockerService{containersErr: apperrors.ErrNotFound}
	r := newDockerTestRouter(svc)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/containers", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", w.Code)
	}
}

func TestDockerHandler_StartContainer_NotFound(t *testing.T) {
	svc := &stubDockerService{startErr: apperrors.ErrNotFound}
	r := newDockerTestRouter(svc)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/containers/missing/start", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", w.Code)
	}
}

func TestDockerHandler_StartContainer_Success(t *testing.T) {
	svc := &stubDockerService{}
	r := newDockerTestRouter(svc)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/containers/abc/start", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", w.Code)
	}
	if svc.startedID != "abc" {
		t.Fatalf("startedID = %q", svc.startedID)
	}
}