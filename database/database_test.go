package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestConnectReturnsErrorForInvalidDSN(t *testing.T) {
	_, err := Connect("://invalid")
	if err == nil {
		t.Fatal("Expected invalid DSN error")
	}
}

func TestMigrate(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sql.OpenDB(noopConnector{})}), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("Expected dry run database to open: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("Expected dry run migration to succeed: %v", err)
	}
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
	return &noopRows{}, nil
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

type noopRows struct {
	sent bool
}

func (*noopRows) Columns() []string {
	return []string{"count"}
}

func (*noopRows) Close() error {
	return nil
}

func (r *noopRows) Next(dest []driver.Value) error {
	if r.sent {
		return io.EOF
	}
	r.sent = true
	dest[0] = int64(0)
	return nil
}
