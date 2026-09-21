// Package setup_test contains test-only constructors and helpers for the
// services package. The `setup_test` package name makes this directory
// only importable from `*_test.go` files, so production binaries never
// pull it in.
//
// Tests import it as:
//
//	import setup "server-monitoring/test/setup"
//
// and call e.g. `setup.NewServerServiceWithRepo(fakeRepo)`.
package setup_test

import (
	"context"
	"fmt"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	"server-monitoring/internal/repository"
	"server-monitoring/internal/services"
	"server-monitoring/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ---- Service constructors (repo-backed variants) --------------------------------

// NewServerServiceWithRepo builds a ServerService backed by the given
// repository.
func NewServerServiceWithRepo(repo repository.ServerRepository) services.ServerService {
	return services.NewServerServiceFromRepo(repo)
}

// NewUserServiceWithRepo builds a UserService backed by the given
// repository.
func NewUserServiceWithRepo(repo repository.UserRepository) services.UserService {
	return services.NewUserServiceFromRepo(repo)
}

// NewStateServiceWithRepo builds a StateService backed by the given
// repository.
func NewStateServiceWithRepo(repo repository.StateRepository) services.StateService {
	return services.NewStateServiceFromRepo(repo)
}

// NewHistoryServiceWithRepo builds a HistoryService backed by the given
// repository.
func NewHistoryServiceWithRepo(repo repository.HistoryRepository, retention time.Duration) services.HistoryService {
	return services.NewHistoryServiceFromRepo(repo, retention)
}

// ---- PM2 -----------------------------------------------------------------------

// NewPM2ServiceWithDialErr builds a PM2Service that surfaces the given
// dial error as a 503 on every call.
func NewPM2ServiceWithDialErr(err error) services.PM2Service {
	return services.NewPM2ServiceFromDialErr(err)
}

// ---- Agent ----------------------------------------------------------------------

// AgentIdentityChanged reports whether the push carries a different
// identity from the stored server. Empty host/port from the push are
// treated as "no opinion".
func AgentIdentityChanged(server models.Server, push dtos.AgentPushDTO) bool {
	return services.IdentityChanged(server, push)
}

// AgentTouchIdentity updates the mutable CRUD fields the agent sent
// without overwriting LastSeen / Status / AgentToken / CreatedAt.
func AgentTouchIdentity(ctx context.Context, repo repository.ServerRepository, id bson.ObjectID, push dtos.AgentPushDTO) error {
	return services.TouchIdentity(ctx, repo, id, push)
}

// ---- Docker / scheduler helpers -------------------------------------------------

// FormatBytes renders an int64 byte count as a human-readable string.
func FormatBytes(size int64) string {
	return services.FormatBytesHuman(size)
}

// SnapshotsFromPM2Processes builds the ServiceSnapshot list for one
// tick of PM2 processes on the given server.
func SnapshotsFromPM2Processes(serverID string, procs []models.PM2Process) []models.ServiceSnapshot {
	return services.SnapshotsFromPM2(serverID, procs)
}

// SnapshotsFromDockerContainers builds the ServiceSnapshot list for one
// tick of Docker containers on the given server.
func SnapshotsFromDockerContainers(serverID string, containers []models.DockerContainer) []models.ServiceSnapshot {
	return services.SnapshotsFromDocker(serverID, containers)
}

func ServerIDFromQuery(c *gin.Context, serverSvc services.ServerService) string {
	return utils.ServerIDFromQuery(c, serverSvc)

}

// silence unused import warnings if a helper is dropped.
var _ = fmt.Sprintf
