package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/middlewares"
	"server-monitoring/internal/models"
	"server-monitoring/internal/services"

	jwtutil "server-monitoring/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func newAuthTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", middlewares.AuthMiddleware(), func(c *gin.Context) {
		uid, _ := c.Get("userID")
		c.JSON(http.StatusOK, gin.H{"userID": uid})
	})
	return r
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	r := newAuthTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_WrongScheme(t *testing.T) {
	r := newAuthTestRouter()
	cases := []string{"", "Basic abc", "Bearer", "Bearer ", "bearer abc"}
	for _, h := range cases {
		t.Run(h, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/x", nil)
			if h != "" {
				req.Header.Set("Authorization", h)
			}
			r.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("code = %d, want 401", w.Code)
			}
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	r := newAuthTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer not-a-jwt")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Invalid or expired") {
		t.Fatalf("body = %s, want invalid token message", w.Body.String())
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	tok, err := jwtutil.GenerateToken("user-42")
	if err != nil {
		t.Fatal(err)
	}
	r := newAuthTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "user-42") {
		t.Fatalf("body = %s, want user-42", w.Body.String())
	}
}

// fakeServerService resolves tokens against an in-memory map.
type fakeServerService struct {
	tokens map[string]*models.Server
}

func (s *fakeServerService) ResolveByAgentToken(_ context.Context, token string) (*models.Server, error) {
	if srv, ok := s.tokens[token]; ok {
		return srv, nil
	}
	return nil, nil
}

func (s *fakeServerService) Create(context.Context, dtos.CreateServerDTO) (*models.Server, error) {
	return nil, nil
}
func (s *fakeServerService) List(context.Context) ([]models.Server, error) { return nil, nil }
func (s *fakeServerService) GetByID(context.Context, bson.ObjectID) (*models.Server, error) {
	return nil, nil
}
func (s *fakeServerService) Update(context.Context, bson.ObjectID, dtos.UpdateServerDTO) (*models.Server, error) {
	return nil, nil
}
func (s *fakeServerService) Delete(context.Context, bson.ObjectID) error { return nil }
func (s *fakeServerService) GenerateAgentToken(context.Context, bson.ObjectID) (*models.Server, string, error) {
	return nil, "", nil
}
func (s *fakeServerService) TouchSeen(context.Context, bson.ObjectID, models.ServerStatus) error {
	return nil
}
func (s *fakeServerService) EnsureLocalServer(context.Context) (*models.Server, error) {
	return nil, nil
}

func newAgentTestRouter(svc services.ServerService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", middlewares.AgentAuthMiddleware(svc), func(c *gin.Context) {
		srv := middlewares.AgentFromContext(c)
		if srv == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "missing"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": srv.HexID()})
	})
	return r
}

func TestAgentAuth_MissingHeader(t *testing.T) {
	svc := &fakeServerService{tokens: map[string]*models.Server{}}
	r := newAgentTestRouter(svc)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", w.Code)
	}
}

func TestAgentAuth_EmptyToken(t *testing.T) {
	svc := &fakeServerService{tokens: map[string]*models.Server{}}
	r := newAgentTestRouter(svc)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer ")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", w.Code)
	}
}

func TestAgentAuth_WrongScheme(t *testing.T) {
	svc := &fakeServerService{tokens: map[string]*models.Server{}}
	r := newAgentTestRouter(svc)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Basic abc")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", w.Code)
	}
}

func TestAgentAuth_UnknownToken(t *testing.T) {
	svc := &fakeServerService{tokens: map[string]*models.Server{}}
	r := newAgentTestRouter(svc)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer nope")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", w.Code)
	}
}

func TestAgentAuth_ValidToken(t *testing.T) {
	id := bson.NewObjectID()
	svc := &fakeServerService{tokens: map[string]*models.Server{
		"good-token": {ID: id},
	}}
	r := newAgentTestRouter(svc)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200; body = %s", w.Code, w.Body.String())
	}
}

func TestAgentFromContext_NilContextReturnsNil(t *testing.T) {
	// AgentFromContext guards against a missing key (returns nil), but
	// panics on a nil *gin.Context — so we test the missing-key path
	// only.
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if got := middlewares.AgentFromContext(c); got != nil {
		t.Fatalf("got %+v, want nil", got)
	}
}