package repositories

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/TB-Systems/tb-project-manager-api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestNewAuthRepository(t *testing.T) {
	repository := NewAuthRepository(nil)
	if repository == nil {
		t.Fatal("Expected auth repository to be initialized")
	}
}

func TestAuthRepositoryDryRunQueries(t *testing.T) {
	db := dryRunDB(t)
	repository := NewAuthRepository(db)
	ctx := context.Background()

	t.Run("authenticate returns error for empty dry run user password", func(t *testing.T) {
		_, err := repository.Authenticate(ctx, models.Login{Username: "tiago", Password: "password"})
		if err == nil {
			t.Fatal("Expected bcrypt error for empty dry run password")
		}
	})

	t.Run("find valid session builds query", func(t *testing.T) {
		_, err := repository.FindValidSessionByTokenHash(ctx, "hash", time.Now().UTC())
		if err != nil {
			t.Fatalf("Expected dry run session lookup to succeed, got %v", err)
		}
	})

	t.Run("revoke session builds update", func(t *testing.T) {
		err := repository.RevokeSessionByTokenHash(ctx, "hash", time.Now().UTC())
		if err != nil {
			t.Fatalf("Expected dry run revoke to succeed, got %v", err)
		}
	})

	t.Run("touch session builds update", func(t *testing.T) {
		err := repository.TouchSession(ctx, 1, time.Now().UTC())
		if err != nil {
			t.Fatalf("Expected dry run touch to succeed, got %v", err)
		}
	})

	t.Run("upsert session builds create", func(t *testing.T) {
		err := repository.UpsertSession(ctx, models.UserSession{
			UserID:     1,
			TokenHash:  "hash",
			ExpiresAt:  time.Now().Add(time.Hour),
			LastSeenAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("Expected dry run upsert to succeed, got %v", err)
		}
	})
}

func dryRunDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sql.OpenDB(noopConnector{})}), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("Expected dry run database to open: %v", err)
	}

	return db
}

type noopConnector struct{}

func (noopConnector) Connect(context.Context) (driver.Conn, error) {
	return noopConn{}, nil
}

func (noopConnector) Driver() driver.Driver {
	return noopDriver{}
}

type noopDriver struct{}

func (noopDriver) Open(string) (driver.Conn, error) {
	return noopConn{}, nil
}

type noopConn struct{}

func (noopConn) Prepare(string) (driver.Stmt, error) {
	return noopStmt{}, nil
}

func (noopConn) Close() error {
	return nil
}

func (noopConn) Begin() (driver.Tx, error) {
	return noopTx{}, nil
}

type noopStmt struct{}

func (noopStmt) Close() error {
	return nil
}

func (noopStmt) NumInput() int {
	return -1
}

func (noopStmt) Exec([]driver.Value) (driver.Result, error) {
	return noopResult{}, nil
}

func (noopStmt) Query([]driver.Value) (driver.Rows, error) {
	return noopRows{}, nil
}

type noopTx struct{}

func (noopTx) Commit() error {
	return nil
}

func (noopTx) Rollback() error {
	return nil
}

type noopResult struct{}

func (noopResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (noopResult) RowsAffected() (int64, error) {
	return 0, nil
}

type noopRows struct{}

func (noopRows) Columns() []string {
	return nil
}

func (noopRows) Close() error {
	return nil
}

func (noopRows) Next([]driver.Value) error {
	return driver.ErrBadConn
}
