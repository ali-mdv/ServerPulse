package handlers_test

import (
	"context"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	"server-monitoring/internal/services"

	"github.com/docker/docker/api/types/container"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// stubAuthService implements services.AuthService.
type stubAuthService struct {
	token string
	err   error
	got   dtos.LoginDTO
}

func (s *stubAuthService) GenerateAccessToken(dto dtos.LoginDTO) (string, error) {
	s.got = dto
	return s.token, s.err
}

// stubUserService implements services.UserService.
type stubUserService struct {
	list        []models.User
	listErr     error
	getByIDUser *models.User
	getByIDErr  error
	createUser  *models.User
	createErr   error
	updatedUser *models.User
	updateErr   error
}

func (s *stubUserService) UsersList() (*[]models.User, error) { return &s.list, s.listErr }
func (s *stubUserService) FindUserByID(string) (*models.User, error) {
	return s.getByIDUser, s.getByIDErr
}
func (s *stubUserService) FindUserByEmail(string) (*models.User, error) { return nil, nil }
func (s *stubUserService) CreateUser(dtos.CreateUserDTO) (*models.User, error) {
	return s.createUser, s.createErr
}
func (s *stubUserService) UpdateUserByID(string, dtos.UpdateUserDTO) (*models.User, error) {
	return s.updatedUser, s.updateErr
}
func (s *stubUserService) GenerateHash(string) (*string, error) { return nil, nil }

// stubServerSvc implements services.ServerService for handler tests.
type stubServerSvc struct {
	list       []models.Server
	listErr    error
	getByID    *models.Server
	getByIDErr error
	created    *models.Server
	createErr  error
	updated    *models.Server
	updateErr  error
	deleted    error
	genToken   string
	genErr     error
}

func (s *stubServerSvc) Create(context.Context, dtos.CreateServerDTO) (*models.Server, error) {
	return s.created, s.createErr
}
func (s *stubServerSvc) List(context.Context) ([]models.Server, error) { return s.list, s.listErr }
func (s *stubServerSvc) GetByID(context.Context, bson.ObjectID) (*models.Server, error) {
	return s.getByID, s.getByIDErr
}
func (s *stubServerSvc) Update(context.Context, bson.ObjectID, dtos.UpdateServerDTO) (*models.Server, error) {
	return s.updated, s.updateErr
}
func (s *stubServerSvc) Delete(context.Context, bson.ObjectID) error { return s.deleted }
func (s *stubServerSvc) GenerateAgentToken(context.Context, bson.ObjectID) (*models.Server, string, error) {
	if s.genErr != nil {
		return nil, "", s.genErr
	}
	return s.updated, s.genToken, nil
}
func (s *stubServerSvc) TouchSeen(context.Context, bson.ObjectID, models.ServerStatus) error { return nil }
func (s *stubServerSvc) ResolveByAgentToken(context.Context, string) (*models.Server, error) {
	return nil, nil
}
func (s *stubServerSvc) EnsureLocalServer(context.Context) (*models.Server, error) { return nil, nil }

// stubStateSvc implements services.StateService for the server handler.
type stubStateSvc struct {
	usage    *models.ServerUsageState
	usageErr error
	pm2      *models.ProviderState
	docker   *models.ProviderState
	stateErr error
}

func (s *stubStateSvc) RecordProviderState(context.Context, string, string, bool, []models.ServiceSnapshot) error {
	return nil
}
func (s *stubStateSvc) RecordServerUsage(context.Context, string, models.SystemUsage) error { return nil }
func (s *stubStateSvc) GetProviderState(context.Context, string, string) (*models.ProviderState, error) {
	return s.pm2, s.stateErr
}
func (s *stubStateSvc) GetServerUsage(context.Context, string) (*models.ServerUsageState, error) {
	return s.usage, s.usageErr
}

// minimalStateServerSvc is the bare ServerService stub used by state
// handler tests — only EnsureLocalServer is invoked.
type minimalStateServerSvc struct{ local *models.Server }

func (s *minimalStateServerSvc) Create(context.Context, dtos.CreateServerDTO) (*models.Server, error) {
	return nil, nil
}
func (s *minimalStateServerSvc) List(context.Context) ([]models.Server, error) { return nil, nil }
func (s *minimalStateServerSvc) GetByID(context.Context, bson.ObjectID) (*models.Server, error) {
	return nil, nil
}
func (s *minimalStateServerSvc) Update(context.Context, bson.ObjectID, dtos.UpdateServerDTO) (*models.Server, error) {
	return nil, nil
}
func (s *minimalStateServerSvc) Delete(context.Context, bson.ObjectID) error { return nil }
func (s *minimalStateServerSvc) GenerateAgentToken(context.Context, bson.ObjectID) (*models.Server, string, error) {
	return nil, "", nil
}
func (s *minimalStateServerSvc) TouchSeen(context.Context, bson.ObjectID, models.ServerStatus) error {
	return nil
}
func (s *minimalStateServerSvc) ResolveByAgentToken(context.Context, string) (*models.Server, error) {
	return nil, nil
}
func (s *minimalStateServerSvc) EnsureLocalServer(context.Context) (*models.Server, error) {
	return s.local, nil
}

// stubStateService2 is the variant used by the state handler that
// branches between pm2 and docker.
type stubStateService2 struct {
	pm2      *models.ProviderState
	docker   *models.ProviderState
	stateErr error
	usage    *models.ServerUsageState
	usageErr error
}

func (s *stubStateService2) RecordProviderState(context.Context, string, string, bool, []models.ServiceSnapshot) error {
	return nil
}
func (s *stubStateService2) RecordServerUsage(context.Context, string, models.SystemUsage) error {
	return nil
}
func (s *stubStateService2) GetProviderState(_ context.Context, _, provider string) (*models.ProviderState, error) {
	if s.stateErr != nil {
		return nil, s.stateErr
	}
	if provider == "pm2" {
		return s.pm2, nil
	}
	return s.docker, nil
}
func (s *stubStateService2) GetServerUsage(context.Context, string) (*models.ServerUsageState, error) {
	return s.usage, s.usageErr
}

// stubDockerService implements services.DockerService.
type stubDockerService struct {
	containers     []models.DockerContainer
	containersErr error
	startedID     string
	startErr      error
}

func (s *stubDockerService) ImagesList(bool) ([]models.DockerImage, error) { return nil, nil }
func (s *stubDockerService) ContainersList(bool) ([]models.DockerContainer, error) {
	return s.containers, s.containersErr
}
func (s *stubDockerService) InspectContainer(string) (container.InspectResponse, error) {
	return container.InspectResponse{}, nil
}
func (s *stubDockerService) StartContainer(id string) error {
	s.startedID = id
	return s.startErr
}
func (s *stubDockerService) StopContainer(string) error   { return nil }
func (s *stubDockerService) RestartContainer(string) error { return nil }
func (s *stubDockerService) FetchContainerLogs(string, int) (string, error) {
	return "", nil
}

// stubPM2Service implements services.PM2Service.
type stubPM2Service struct {
	list       []models.PM2Process
	listErr    error
	proc       *models.PM2Process
	procErr    error
	startErr   error
	stopErr    error
	restartErr error
	logsErr    error
	logs       string
}

func (s *stubPM2Service) List() ([]models.PM2Process, error) { return s.list, s.listErr }
func (s *stubPM2Service) FindPM2ProcessByID(int) (*models.PM2Process, error) {
	return s.proc, s.procErr
}
func (s *stubPM2Service) StartPM2ProcessByID(int) error   { return s.startErr }
func (s *stubPM2Service) StopPM2ProcessByID(int) error    { return s.stopErr }
func (s *stubPM2Service) RestartPM2ProcessByID(int) error { return s.restartErr }
func (s *stubPM2Service) FetchContainerLogs(int, int) (string, error) {
	return s.logs, s.logsErr
}

// minimalHistoryServerSvc is the bare ServerService stub used by
// history handler tests.
type minimalHistoryServerSvc struct{}

func (s *minimalHistoryServerSvc) Create(context.Context, dtos.CreateServerDTO) (*models.Server, error) {
	return nil, nil
}
func (s *minimalHistoryServerSvc) List(context.Context) ([]models.Server, error) { return nil, nil }
func (s *minimalHistoryServerSvc) GetByID(context.Context, bson.ObjectID) (*models.Server, error) {
	return nil, nil
}
func (s *minimalHistoryServerSvc) Update(context.Context, bson.ObjectID, dtos.UpdateServerDTO) (*models.Server, error) {
	return nil, nil
}
func (s *minimalHistoryServerSvc) Delete(context.Context, bson.ObjectID) error { return nil }
func (s *minimalHistoryServerSvc) GenerateAgentToken(context.Context, bson.ObjectID) (*models.Server, string, error) {
	return nil, "", nil
}
func (s *minimalHistoryServerSvc) TouchSeen(context.Context, bson.ObjectID, models.ServerStatus) error {
	return nil
}
func (s *minimalHistoryServerSvc) ResolveByAgentToken(context.Context, string) (*models.Server, error) {
	return nil, nil
}
func (s *minimalHistoryServerSvc) EnsureLocalServer(context.Context) (*models.Server, error) {
	return nil, nil
}

// stubHistoryService implements services.HistoryService.
type stubHistoryService struct {
	tracked    []models.SnapshotMeta
	trackedErr error
	points     []models.ServiceSnapshot
	pointsErr  error
}

func (s *stubHistoryService) EnsureSchema(context.Context) error { return nil }
func (s *stubHistoryService) RecordPM2Snapshot(context.Context, string, []models.PM2Process) error {
	return nil
}
func (s *stubHistoryService) RecordDockerSnapshot(context.Context, string, []models.DockerContainer) error {
	return nil
}
func (s *stubHistoryService) RecordSystemSnapshot(context.Context, string, models.SystemUsage) error {
	return nil
}
func (s *stubHistoryService) RecordProviderUnavailable(context.Context, string, string) error {
	return nil
}
func (s *stubHistoryService) InsertHistoryBatch(context.Context, []models.ServiceSnapshot) error {
	return nil
}
func (s *stubHistoryService) ListTrackedServices(context.Context, string, string) ([]models.SnapshotMeta, error) {
	return s.tracked, s.trackedErr
}
func (s *stubHistoryService) FindSeries(context.Context, string, string, string, time.Time, time.Time, time.Duration) ([]models.ServiceSnapshot, error) {
	return s.points, s.pointsErr
}

// idleServerSvc is the bare ServerService stub used by server_id tests.
type idleServerSvc struct{ local *models.Server }

func (s *idleServerSvc) Create(context.Context, dtos.CreateServerDTO) (*models.Server, error) {
	return nil, nil
}
func (s *idleServerSvc) List(context.Context) ([]models.Server, error) { return nil, nil }
func (s *idleServerSvc) GetByID(context.Context, bson.ObjectID) (*models.Server, error) {
	return nil, nil
}
func (s *idleServerSvc) Update(context.Context, bson.ObjectID, dtos.UpdateServerDTO) (*models.Server, error) {
	return nil, nil
}
func (s *idleServerSvc) Delete(context.Context, bson.ObjectID) error { return nil }
func (s *idleServerSvc) GenerateAgentToken(context.Context, bson.ObjectID) (*models.Server, string, error) {
	return nil, "", nil
}
func (s *idleServerSvc) TouchSeen(context.Context, bson.ObjectID, models.ServerStatus) error {
	return nil
}
func (s *idleServerSvc) ResolveByAgentToken(context.Context, string) (*models.Server, error) {
	return nil, nil
}
func (s *idleServerSvc) EnsureLocalServer(context.Context) (*models.Server, error) {
	return s.local, nil
}

// ensure interfaces stay satisfied.
var (
	_ services.AuthService    = (*stubAuthService)(nil)
	_ services.UserService    = (*stubUserService)(nil)
	_ services.ServerService  = (*stubServerSvc)(nil)
	_ services.StateService   = (*stubStateSvc)(nil)
	_ services.StateService   = (*stubStateService2)(nil)
	_ services.ServerService  = (*minimalStateServerSvc)(nil)
	_ services.ServerService  = (*minimalHistoryServerSvc)(nil)
	_ services.HistoryService = (*stubHistoryService)(nil)
	_ services.ServerService  = (*idleServerSvc)(nil)
	_ services.DockerService  = (*stubDockerService)(nil)
	_ services.PM2Service     = (*stubPM2Service)(nil)
)