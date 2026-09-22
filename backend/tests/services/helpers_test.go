package services_test

import (
	"context"
	"errors"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	"server-monitoring/internal/repository"
	apperrors "server-monitoring/pkg/errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

// fakeServerRepo is the in-memory repository used by services/server_*
// tests. It implements repository.ServerRepository against a slice.
type fakeServerRepo struct {
	servers []models.Server
}

func newFakeServerRepo() *fakeServerRepo { return &fakeServerRepo{} }

func (r *fakeServerRepo) Insert(_ context.Context, s models.Server) (*models.Server, error) {
	if s.ID.IsZero() {
		s.ID = bson.NewObjectID()
	}
	s.CreatedAt = time.Now().UTC()
	s.UpdatedAt = s.CreatedAt
	r.servers = append(r.servers, s)
	return &s, nil
}

func (r *fakeServerRepo) FindByID(_ context.Context, id bson.ObjectID) (*models.Server, error) {
	for i := range r.servers {
		if r.servers[i].ID == id {
			return &r.servers[i], nil
		}
	}
	return nil, nil
}

func (r *fakeServerRepo) FindByName(_ context.Context, name string) (*models.Server, error) {
	for i := range r.servers {
		if r.servers[i].Name == name {
			return &r.servers[i], nil
		}
	}
	return nil, nil
}

func (r *fakeServerRepo) FindByAgentToken(_ context.Context, token string) (*models.Server, error) {
	for i := range r.servers {
		if r.servers[i].AgentToken == token {
			return &r.servers[i], nil
		}
	}
	return nil, nil
}

func (r *fakeServerRepo) List(_ context.Context) ([]models.Server, error) {
	return r.servers, nil
}

func (r *fakeServerRepo) UpdateByID(_ context.Context, id bson.ObjectID, update bson.D) (*models.Server, error) {
	for i := range r.servers {
		if r.servers[i].ID != id {
			continue
		}
		for _, top := range update {
			if top.Key != "$set" {
				continue
			}
			sets, ok := top.Value.(bson.D)
			if !ok {
				continue
			}
			for _, set := range sets {
				switch set.Key {
				case "name":
					r.servers[i].Name = set.Value.(string)
				case "host":
					r.servers[i].Host = set.Value.(string)
				case "port":
					r.servers[i].Port = set.Value.(int)
				case "description":
					r.servers[i].Description = set.Value.(string)
				case "agentToken":
					r.servers[i].AgentToken = set.Value.(string)
				case "updatedAt":
					r.servers[i].UpdatedAt = set.Value.(time.Time)
				}
			}
		}
		return &r.servers[i], nil
	}
	return nil, nil
}

func (r *fakeServerRepo) DeleteByID(context.Context, bson.ObjectID) error { return nil }
func (r *fakeServerRepo) TouchSeen(context.Context, bson.ObjectID, models.ServerStatus) error {
	return nil
}

// fakeUserRepo is the in-memory repository used by services/user_service_test.
type fakeUserRepo struct {
	users map[string]*models.User
}

func newFakeUserRepo() *fakeUserRepo { return &fakeUserRepo{users: map[string]*models.User{}} }

func (r *fakeUserRepo) UsersList() (*[]models.User, error) {
	out := make([]models.User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, *u)
	}
	return &out, nil
}

func (r *fakeUserRepo) FindUserByID(userID string) (*models.User, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("bad id")
	}
	for _, u := range r.users {
		if u.ID == oid {
			return u, nil
		}
	}
	return nil, mongo.ErrNoDocuments
}

func (r *fakeUserRepo) FindUserByEmail(email string) (*models.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, mongo.ErrNoDocuments
}

func (r *fakeUserRepo) CreateUser(dto dtos.CreateUserDTO) (*models.User, error) {
	for _, u := range r.users {
		if u.Email == dto.Email {
			return nil, mongo.CommandError{Code: 11000}
		}
	}
	u := &models.User{
		ID:       bson.NewObjectID(),
		Email:    dto.Email,
		Password: dto.Password,
	}
	r.users[u.Email] = u
	return u, nil
}

func (r *fakeUserRepo) UpdateUserByID(userID string, data dtos.UpdateUserDTO) (*models.User, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("bad id")
	}
	var target *models.User
	for _, u := range r.users {
		if u.ID == oid {
			target = u
			break
		}
	}
	if target == nil {
		return nil, mongo.ErrNoDocuments
	}
	if data.Email != nil {
		target.Email = *data.Email
	}
	if data.Password != nil {
		target.Password = *data.Password
	}
	return target, nil
}

// fakeStateRepo is the in-memory repository used by services/state_service_test.
type fakeStateRepo struct {
	providerStates []models.ProviderState
	serverUsages   []models.ServerUsageState
}

func (r *fakeStateRepo) UpsertProviderState(_ context.Context, s models.ProviderState) error {
	for i := range r.providerStates {
		if r.providerStates[i].ServerID == s.ServerID && r.providerStates[i].Provider == s.Provider {
			r.providerStates[i] = s
			return nil
		}
	}
	r.providerStates = append(r.providerStates, s)
	return nil
}

func (r *fakeStateRepo) UpsertServerUsage(_ context.Context, s models.ServerUsageState) error {
	for i := range r.serverUsages {
		if r.serverUsages[i].ServerID == s.ServerID {
			r.serverUsages[i] = s
			return nil
		}
	}
	r.serverUsages = append(r.serverUsages, s)
	return nil
}

func (r *fakeStateRepo) GetProviderState(_ context.Context, serverID, provider string) (*models.ProviderState, error) {
	for i := range r.providerStates {
		if r.providerStates[i].ServerID == serverID && r.providerStates[i].Provider == provider {
			s := r.providerStates[i]
			return &s, nil
		}
	}
	return nil, nil
}

func (r *fakeStateRepo) GetServerUsage(_ context.Context, serverID string) (*models.ServerUsageState, error) {
	for i := range r.serverUsages {
		if r.serverUsages[i].ServerID == serverID {
			s := r.serverUsages[i]
			return &s, nil
		}
	}
	return nil, nil
}

// fakeHistoryRepo is the in-memory repository used by services/history_service_test.
type fakeHistoryRepo struct {
	batches    [][]models.ServiceSnapshot
	ensured    int
	tracked    []models.SnapshotMeta
	retentions []time.Duration
}

func (r *fakeHistoryRepo) EnsureCollection(_ context.Context, retention time.Duration) error {
	r.ensured++
	r.retentions = append(r.retentions, retention)
	return nil
}
func (r *fakeHistoryRepo) InsertBatch(_ context.Context, b []models.ServiceSnapshot) error {
	cp := make([]models.ServiceSnapshot, len(b))
	copy(cp, b)
	r.batches = append(r.batches, cp)
	return nil
}
func (r *fakeHistoryRepo) ListTrackedServices(context.Context, string, string) ([]models.SnapshotMeta, error) {
	return r.tracked, nil
}
func (r *fakeHistoryRepo) FindSeries(context.Context, string, string, string, time.Time, time.Time, time.Duration) ([]models.ServiceSnapshot, error) {
	return nil, nil
}
func (r *fakeHistoryRepo) FindAvailability(context.Context, string, string, time.Time, time.Time) ([]models.ServiceSnapshot, error) {
	return nil, nil
}

// fakeUserService implements UserService for auth-service tests.
type fakeUserService struct {
	users map[string]*models.User
}

func newFakeUserService() *fakeUserService { return &fakeUserService{users: map[string]*models.User{}} }

func (s *fakeUserService) UsersList() (*[]models.User, error) {
	out := make([]models.User, 0, len(s.users))
	for _, u := range s.users {
		out = append(out, *u)
	}
	return &out, nil
}
func (s *fakeUserService) FindUserByID(string) (*models.User, error) { return nil, nil }
func (s *fakeUserService) FindUserByEmail(email string) (*models.User, error) {
	if u, ok := s.users[email]; ok {
		return u, nil
	}
	return nil, apperrors.ErrNotFound
}
func (s *fakeUserService) CreateUser(dtos.CreateUserDTO) (*models.User, error) { return nil, nil }
func (s *fakeUserService) UpdateUserByID(string, dtos.UpdateUserDTO) (*models.User, error) {
	return nil, nil
}
func (s *fakeUserService) GenerateHash(p string) (*string, error) {
	h := hashForTest(p)
	return &h, nil
}

// fakeAgentRepo is the minimum surface ServerRepository exposes that the
// agent service uses.
type fakeAgentRepo struct {
	servers     []models.Server
	updateCalls int
}

func (r *fakeAgentRepo) Insert(context.Context, models.Server) (*models.Server, error) {
	return nil, nil
}
func (r *fakeAgentRepo) FindByID(context.Context, bson.ObjectID) (*models.Server, error) {
	return nil, nil
}
func (r *fakeAgentRepo) FindByName(context.Context, string) (*models.Server, error) { return nil, nil }
func (r *fakeAgentRepo) FindByAgentToken(context.Context, string) (*models.Server, error) {
	return nil, nil
}
func (r *fakeAgentRepo) List(context.Context) ([]models.Server, error) { return nil, nil }
func (r *fakeAgentRepo) UpdateByID(_ context.Context, id bson.ObjectID, _ bson.D) (*models.Server, error) {
	r.updateCalls++
	for i := range r.servers {
		if r.servers[i].ID == id {
			return &r.servers[i], nil
		}
	}
	return nil, nil
}
func (r *fakeAgentRepo) DeleteByID(context.Context, bson.ObjectID) error { return nil }
func (r *fakeAgentRepo) TouchSeen(_ context.Context, id bson.ObjectID, status models.ServerStatus) error {
	for i := range r.servers {
		if r.servers[i].ID == id {
			now := time.Now().UTC()
			r.servers[i].Status = status
			r.servers[i].LastSeen = &now
			return nil
		}
	}
	return nil
}

// fakeAgentState records calls into StateService during agent ingestion.
type fakeAgentState struct {
	providerWrites []fakeStateWrite
	usageWrites    []models.ServerUsageState
}

type fakeStateWrite struct {
	ServerID  string
	Provider  string
	Available bool
	Services  []models.ServiceSnapshot
}

func (s *fakeAgentState) RecordProviderState(_ context.Context, serverID, provider string, available bool, snapshots []models.ServiceSnapshot) error {
	s.providerWrites = append(s.providerWrites, fakeStateWrite{
		ServerID: serverID, Provider: provider, Available: available, Services: snapshots,
	})
	return nil
}
func (s *fakeAgentState) RecordServerUsage(_ context.Context, serverID string, u models.SystemUsage) error {
	s.usageWrites = append(s.usageWrites, models.ServerUsageState{ServerID: serverID, Usage: u})
	return nil
}
func (s *fakeAgentState) GetProviderState(context.Context, string, string) (*models.ProviderState, error) {
	return nil, nil
}
func (s *fakeAgentState) GetServerUsage(context.Context, string) (*models.ServerUsageState, error) {
	return nil, nil
}

// fakeAgentHistory records history inserts.
type fakeAgentHistory struct {
	batches [][]models.ServiceSnapshot
}

func (h *fakeAgentHistory) InsertHistoryBatch(_ context.Context, b []models.ServiceSnapshot) error {
	h.batches = append(h.batches, b)
	return nil
}
func (h *fakeAgentHistory) RecordSystemSnapshot(context.Context, string, models.SystemUsage) error {
	return nil
}
func (h *fakeAgentHistory) RecordPM2Snapshot(context.Context, string, []models.PM2Process) error {
	return nil
}
func (h *fakeAgentHistory) RecordDockerSnapshot(context.Context, string, []models.DockerContainer) error {
	return nil
}
func (h *fakeAgentHistory) RecordProviderUnavailable(context.Context, string, string) error {
	return nil
}
func (h *fakeAgentHistory) EnsureSchema(context.Context) error { return nil }
func (h *fakeAgentHistory) ListTrackedServices(context.Context, string, string) ([]models.SnapshotMeta, error) {
	return nil, nil
}
func (h *fakeAgentHistory) FindSeries(context.Context, string, string, string, time.Time, time.Time, time.Duration) ([]models.ServiceSnapshot, error) {
	return nil, nil
}

// fakeNotificationRepo is the in-memory repository used by
// services/notification_service_test.
type fakeNotificationRepo struct {
	items []models.Notification
}

func (r *fakeNotificationRepo) EnsureIndexes(context.Context) error { return nil }

func (r *fakeNotificationRepo) Insert(_ context.Context, n models.Notification) error {
	r.items = append(r.items, n)
	return nil
}

func (r *fakeNotificationRepo) List(_ context.Context, serverID string, severity models.NotificationSeverity, unreadOnly bool, limit int) ([]models.Notification, error) {
	out := []models.Notification{}
	for _, n := range r.items {
		if serverID != "" && n.ServerID != serverID {
			continue
		}
		if severity != "" && n.Severity != severity {
			continue
		}
		if unreadOnly && n.Read {
			continue
		}
		out = append(out, n)
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *fakeNotificationRepo) MarkRead(_ context.Context, id bson.ObjectID) error {
	for i := range r.items {
		if r.items[i].ID == id {
			r.items[i].Read = true
			return nil
		}
	}
	return mongo.ErrNoDocuments
}

func (r *fakeNotificationRepo) MarkAllRead(_ context.Context, serverID string) (int64, error) {
	var updated int64
	for i := range r.items {
		if serverID != "" && r.items[i].ServerID != serverID {
			continue
		}
		if !r.items[i].Read {
			r.items[i].Read = true
			updated++
		}
	}
	return updated, nil
}

func (r *fakeNotificationRepo) CountUnread(_ context.Context, serverID string) (int64, error) {
	var unread int64
	for i := range r.items {
		if serverID != "" && r.items[i].ServerID != serverID {
			continue
		}
		if !r.items[i].Read {
			unread++
		}
	}
	return unread, nil
}

// fakeSettingsRepo is the in-memory repository used by
// services/settings_service_test.
type fakeSettingsRepo struct {
	settings *models.Settings
}

func (r *fakeSettingsRepo) Get(context.Context) (*models.Settings, error) {
	if r.settings == nil {
		return nil, nil
	}
	cp := *r.settings
	return &cp, nil
}

func (r *fakeSettingsRepo) Ensure(_ context.Context, defaults models.Settings) (*models.Settings, error) {
	if r.settings == nil {
		cp := defaults
		r.settings = &cp
	}
	cp := *r.settings
	return &cp, nil
}

func (r *fakeSettingsRepo) Upsert(_ context.Context, settings models.Settings) (*models.Settings, error) {
	cp := settings
	r.settings = &cp
	return &cp, nil
}

// assert that the test fakes still satisfy the interfaces.
var (
	_ repository.ServerRepository       = (*fakeServerRepo)(nil)
	_ repository.UserRepository         = (*fakeUserRepo)(nil)
	_ repository.StateRepository        = (*fakeStateRepo)(nil)
	_ repository.HistoryRepository      = (*fakeHistoryRepo)(nil)
	_ repository.NotificationRepository = (*fakeNotificationRepo)(nil)
	_ repository.SettingsRepository     = (*fakeSettingsRepo)(nil)
)

// hashForTest returns a bcrypt hash with the cheapest cost so tests stay fast.
func hashForTest(p string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
	return string(h)
}