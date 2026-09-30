package handlers_test

import (
	"net/http/httptest"
	"testing"

	"server-monitoring/internal/models"
	"server-monitoring/internal/services"
	setup_test "server-monitoring/tests/setup"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestServerIDFromQuery_ExplicitIDWins(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/x?serverId=explicit-id", nil)

	id := setup_test.ServerIDFromQuery(c, &idleServerSvc{})
	if id != "explicit-id" {
		t.Fatalf("got %q, want explicit-id", id)
	}
}

func TestServerIDFromQuery_LocalSentinelResolves(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/x?serverId=local", nil)

	realID := bson.NewObjectID()
	id := setup_test.ServerIDFromQuery(c, &idleServerSvc{local: &models.Server{ID: realID}})
	if id != realID.Hex() {
		t.Fatalf("got %q, want %q", id, realID.Hex())
	}
}

func TestServerIDFromQuery_NoQueryResolvesLocal(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/x", nil)

	realID := bson.NewObjectID()
	id := setup_test.ServerIDFromQuery(c, &idleServerSvc{local: &models.Server{ID: realID}})
	if id != realID.Hex() {
		t.Fatalf("got %q, want %q", id, realID.Hex())
	}
}

func TestServerIDFromQuery_LocalFailureFallsBackToSentinel(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/x", nil)

	id := setup_test.ServerIDFromQuery(c, nil)
	if id != services.LocalServerID {
		t.Fatalf("got %q, want %q", id, services.LocalServerID)
	}
}
