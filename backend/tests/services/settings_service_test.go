package services_test

import (
	"context"
	"testing"
	"time"

	"server-monitoring/internal/dtos"
	"server-monitoring/internal/models"
	setup_test "server-monitoring/tests/setup"
)

func TestSettingsService_GetCreatesDefaults(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := setup_test.NewSettingsServiceWithRepo(repo)

	got, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != models.SettingsSingletonID {
		t.Fatalf("id = %q, want %q", got.ID, models.SettingsSingletonID)
	}
	if got.Appearance != models.DefaultAppearance {
		t.Fatalf("appearance = %q, want %q", got.Appearance, models.DefaultAppearance)
	}
	if got.HistoryPollInterval != models.DefaultHistoryPollInterval {
		t.Fatalf("poll interval = %v, want %v", got.HistoryPollInterval, models.DefaultHistoryPollInterval)
	}
	if got.HistoryRetention != models.DefaultHistoryRetention {
		t.Fatalf("retention = %v, want %v", got.HistoryRetention, models.DefaultHistoryRetention)
	}
	if repo.settings == nil {
		t.Fatal("expected the singleton to be persisted")
	}
}

func TestSettingsService_GetIsStable(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := setup_test.NewSettingsServiceWithRepo(repo)

	first, _ := svc.Get(context.Background())
	second, _ := svc.Get(context.Background())
	if first.UpdatedAt != second.UpdatedAt {
		t.Fatalf("Get must not recreate the singleton: %v vs %v", first.UpdatedAt, second.UpdatedAt)
	}
}

func TestSettingsService_UpdatePartialPersists(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := setup_test.NewSettingsServiceWithRepo(repo)
	ctx := context.Background()

	appearance := "dark"
	interval := int64(60)
	retention := int64(24 * 60 * 60)

	got, err := svc.Update(ctx, dtos.UpdateSettingsDTO{
		Appearance:                 &appearance,
		HistoryPollIntervalSeconds: &interval,
		HistoryRetentionSeconds:    &retention,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.Appearance != models.AppearanceDark {
		t.Fatalf("appearance = %q", got.Appearance)
	}
	if got.HistoryPollInterval != 60*time.Second {
		t.Fatalf("poll interval = %v", got.HistoryPollInterval)
	}
	if got.HistoryRetention != 24*time.Hour {
		t.Fatalf("retention = %v", got.HistoryRetention)
	}

	persisted, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if persisted.HistoryPollInterval != 60*time.Second || persisted.Appearance != models.AppearanceDark {
		t.Fatalf("update not persisted: %+v", persisted)
	}
}

func TestSettingsService_UpdateValidates(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := setup_test.NewSettingsServiceWithRepo(repo)
	ctx := context.Background()

	badAppearance := "neon"
	if _, err := svc.Update(ctx, dtos.UpdateSettingsDTO{Appearance: &badAppearance}); err == nil {
		t.Fatal("expected error for invalid appearance")
	}

	tooFast := int64(1)
	if _, err := svc.Update(ctx, dtos.UpdateSettingsDTO{HistoryPollIntervalSeconds: &tooFast}); err == nil {
		t.Fatal("expected error for poll interval below minimum")
	}

	tooShort := int64(60)
	if _, err := svc.Update(ctx, dtos.UpdateSettingsDTO{HistoryRetentionSeconds: &tooShort}); err == nil {
		t.Fatal("expected error for retention below minimum")
	}
}

func TestSettingsService_HistoryAccessors(t *testing.T) {
	repo := &fakeSettingsRepo{}
	svc := setup_test.NewSettingsServiceWithRepo(repo)
	ctx := context.Background()

	interval, err := svc.HistoryPollInterval(ctx)
	if err != nil || interval != models.DefaultHistoryPollInterval {
		t.Fatalf("poll interval = %v, err = %v", interval, err)
	}
	retention, err := svc.HistoryRetention(ctx)
	if err != nil || retention != models.DefaultHistoryRetention {
		t.Fatalf("retention = %v, err = %v", retention, err)
	}
}
