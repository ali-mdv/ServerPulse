package models

import "time"

// Appearance selects the dashboard's colour scheme. It mirrors the
// frontend Theme type so a saved value can be applied directly.
type Appearance string

const (
	AppearanceSystem Appearance = "system"
	AppearanceLight  Appearance = "light"
	AppearanceDark   Appearance = "dark"
)

// Valid reports whether the appearance is one of the known values.
func (a Appearance) Valid() bool {
	switch a {
	case AppearanceSystem, AppearanceLight, AppearanceDark:
		return true
	default:
		return false
	}
}

// SettingsSingletonID is the fixed _id of the single settings document.
// Every read and write targets this key, and writes upsert, so the
// collection never holds more than one row.
const SettingsSingletonID = "global"

// Settings defaults. These apply when the singleton is first created and
// any time a stored value is missing or invalid.
const (
	DefaultAppearance          = AppearanceSystem
	DefaultHistoryPollInterval = 30 * time.Second
	DefaultHistoryRetention    = 7 * 24 * time.Hour
)

// Settings is the application-wide configuration singleton. It replaced
// the HISTORY_POLL_INTERVAL / HISTORY_RETENTION env vars: the scheduler
// reads these DB values at runtime, so changes take effect without a
// restart.
type Settings struct {
	ID                  string        `bson:"_id" json:"id"`
	Appearance          Appearance    `bson:"appearance" json:"appearance"`
	HistoryPollInterval time.Duration `bson:"historyPollInterval" json:"historyPollInterval"`
	HistoryRetention    time.Duration `bson:"historyRetention" json:"historyRetention"`
	UpdatedAt           time.Time     `bson:"updatedAt" json:"updatedAt"`
}

// DefaultSettings returns a fully populated singleton ready to insert.
func DefaultSettings() Settings {
	return Settings{
		ID:                  SettingsSingletonID,
		Appearance:          DefaultAppearance,
		HistoryPollInterval: DefaultHistoryPollInterval,
		HistoryRetention:    DefaultHistoryRetention,
		UpdatedAt:           time.Now().UTC(),
	}
}
