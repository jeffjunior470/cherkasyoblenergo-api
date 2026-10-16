package middleware

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"cherkasyoblenergo-api/internal/config"
	"cherkasyoblenergo-api/internal/database"
)

func TestIsSQLiteBusy_ActualDriverError(t *testing.T) {
	db, err := database.ConnectDB(config.Config{DBName: filepath.Join(t.TempDir(), "locked.db"), LogLevel: "silent"})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()
	require.NoError(t, db.Exec("CREATE TABLE lock_test (id INTEGER PRIMARY KEY)").Error)

	ctx := context.Background()
	first, err := sqlDB.Conn(ctx)
	require.NoError(t, err)
	defer first.Close()
	second, err := sqlDB.Conn(ctx)
	require.NoError(t, err)
	defer second.Close()
	_, err = second.ExecContext(ctx, "PRAGMA busy_timeout=0")
	require.NoError(t, err)
	tx, err := first.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, "INSERT INTO lock_test VALUES (1)")
	require.NoError(t, err)
	_, err = second.ExecContext(ctx, "INSERT INTO lock_test VALUES (2)")
	require.Error(t, err)
	require.True(t, isSQLiteBusy(err))
	require.True(t, isSQLiteBusy(fmt.Errorf("transaction: %w", err)))
	require.False(t, isSQLiteBusy(nil))
	require.False(t, isSQLiteBusy(fmt.Errorf("unrelated error")))
}
