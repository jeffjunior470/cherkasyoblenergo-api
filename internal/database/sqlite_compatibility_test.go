package database

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cherkasyoblenergo-api/internal/config"
	"cherkasyoblenergo-api/internal/models"
)

func TestConnectDB_SettingsOnEveryConnection(t *testing.T) {
	db, err := ConnectDB(config.Config{DBName: filepath.Join(t.TempDir(), "settings.db"), LogLevel: "silent"})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Hold both connections open to ensure the pool cannot reuse the first.
	first, err := sqlDB.Conn(context.Background())
	require.NoError(t, err)
	defer first.Close()
	second, err := sqlDB.Conn(context.Background())
	require.NoError(t, err)
	defer second.Close()
	for _, conn := range []*sql.Conn{first, second} {
		var mode string
		var timeout, foreignKeys int
		require.NoError(t, conn.QueryRowContext(context.Background(), "PRAGMA journal_mode").Scan(&mode))
		require.NoError(t, conn.QueryRowContext(context.Background(), "PRAGMA busy_timeout").Scan(&timeout))
		require.NoError(t, conn.QueryRowContext(context.Background(), "PRAGMA foreign_keys").Scan(&foreignKeys))
		require.Equal(t, "wal", mode)
		require.Equal(t, 5000, timeout)
		require.Equal(t, 1, foreignKeys)
	}
}

func TestConnectDB_LegacyDatesAndReopen(t *testing.T) {
	cfg := config.Config{DBName: filepath.Join(t.TempDir(), "schedule.db"), LogLevel: "silent"}
	db, err := ConnectDB(cfg)
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()
	require.NoError(t, db.AutoMigrate(&models.Schedule{}, &models.IPRateLimit{}))

	// The previous mattn driver stores timestamps in this SQLite text format.
	legacyDate := "2026-10-15 21:30:00+00:00"
	require.NoError(t, db.Exec(`INSERT INTO schedules (news_id, title, date, schedule_date, "1_1") VALUES (?, ?, ?, ?, ?)`,
		101, "Графік на 16 жовтня", legacyDate, "2026-10-16", "08:00-10:00").Error)
	var legacy models.Schedule
	require.NoError(t, db.Where("news_id = ?", 101).First(&legacy).Error)
	expected, err := time.Parse("2006-01-02 15:04:05-07:00", legacyDate)
	require.NoError(t, err)
	require.True(t, legacy.Date.Equal(expected))

	// DATE() interprets offsets in UTC: preserve the existing query semantics.
	var count int64
	require.NoError(t, db.Model(&models.Schedule{}).Where("DATE(date) = ?", "2026-10-15").Count(&count).Error)
	require.EqualValues(t, 1, count)

	newSchedule := models.Schedule{NewsID: 102, Title: "Updated", Date: expected.In(time.FixedZone("Kyiv", 3*60*60)).Add(time.Hour), ScheduleDate: "2026-10-16", Col1_1: "10:00-12:00"}
	require.NoError(t, db.Create(&newSchedule).Error)
	require.NoError(t, sqlDB.Close())

	db, err = ConnectDB(cfg)
	require.NoError(t, err)
	reopened, err := db.DB()
	require.NoError(t, err)
	defer reopened.Close()
	require.NoError(t, db.AutoMigrate(&models.Schedule{}, &models.IPRateLimit{}))
	var schedules []models.Schedule
	require.NoError(t, db.Order("date desc, news_id desc").Find(&schedules).Error)
	require.Len(t, schedules, 2)
	require.Equal(t, 102, schedules[0].NewsID)
	require.True(t, schedules[0].Date.Equal(newSchedule.Date))
	require.Equal(t, "08:00-10:00", schedules[1].Col1_1)
}
