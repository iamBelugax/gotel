package gotel

import (
	"context"
	"database/sql"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// dbTracer provides tracing and metrics for database operations.
type dbTracer struct {
	dbName  string
	dbType  string
	tracer  *Tracer
	metrics *CommonMetrics
}

// NewDBTracer creates a new DBTracer instance for a given database.
func NewDBTracer(tracer *Tracer, metrics *CommonMetrics, dbName, dbType string) *dbTracer {
	return &dbTracer{
		dbName:  dbName,
		dbType:  dbType,
		tracer:  tracer,
		metrics: metrics,
	}
}

// Trace wraps a database query or command execution with tracing and metrics.
func (dt *dbTracer) Trace(ctx context.Context, query string, fn func() error) error {
	ctx, span := dt.tracer.StartSpan(ctx, "db.query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBQueryTextKey.String(query),
			semconv.DBSystemKey.String(dt.dbType),
			semconv.DBNamespaceKey.String(dt.dbName),
		),
	)
	defer span.End()

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	attrs := []attribute.KeyValue{
		attribute.String("db_name", dt.dbName),
		attribute.String("db_type", dt.dbType),
	}
	dt.metrics.DBQueriesTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	dt.metrics.DBQueryDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))

	if err != nil {
		span.WithError(err)
		span.WithStatus(codes.Error, err.Error())
	}

	return err
}

// TracedDB is a wrapper around *sql.DB that transparently adds tracing and metrics
// to queries, execs, prepares and transactions.
type TracedDB struct {
	db     *sql.DB
	tracer *dbTracer
}

// NewTracedDB wraps a sql.DB instance with tracing and metrics capabilities.
func NewTracedDB(db *sql.DB, tracer *dbTracer) *TracedDB {
	return &TracedDB{db: db, tracer: tracer}
}

// QueryContext wraps sql.DB.QueryContext with tracing and metrics.
func (tdb *TracedDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	var (
		err  error
		rows *sql.Rows
	)
	tdb.tracer.Trace(ctx, query, func() error {
		rows, err = tdb.db.QueryContext(ctx, query, args...)
		return err
	})
	return rows, err
}

// QueryRowContext wraps sql.DB.QueryRowContext with tracing.
func (tdb *TracedDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	var row *sql.Row
	tdb.tracer.Trace(ctx, query, func() error {
		row = tdb.db.QueryRowContext(ctx, query, args...)
		return nil
	})
	return row
}

// ExecContext wraps sql.DB.ExecContext with tracing and metrics.
func (tdb *TracedDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	var (
		err    error
		result sql.Result
	)
	tdb.tracer.Trace(ctx, query, func() error {
		result, err = tdb.db.ExecContext(ctx, query, args...)
		return err
	})
	return result, err
}

// PrepareContext wraps sql.DB.PrepareContext with tracing and returns a TracedStmt.
func (tdb *TracedDB) PrepareContext(ctx context.Context, query string) (*TracedStmt, error) {
	var (
		err  error
		stmt *sql.Stmt
	)
	tdb.tracer.Trace(ctx, query, func() error {
		stmt, err = tdb.db.PrepareContext(ctx, query)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &TracedStmt{stmt: stmt, tracer: tdb.tracer, query: query}, nil
}

// BeginTx wraps sql.DB.BeginTx with tracing and returns a TracedTx.
func (tdb *TracedDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*TracedTx, error) {
	var (
		err error
		tx  *sql.Tx
	)
	tdb.tracer.Trace(ctx, "BEGIN", func() error {
		tx, err = tdb.db.BeginTx(ctx, opts)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &TracedTx{tx: tx, tracer: tdb.tracer}, nil
}

// PingContext wraps sql.DB.PingContext with tracing.
func (tdb *TracedDB) PingContext(ctx context.Context) error {
	return tdb.tracer.Trace(ctx, "PING", func() error {
		return tdb.db.PingContext(ctx)
	})
}

// TracedStmt wraps *sql.Stmt with tracing.
type TracedStmt struct {
	query  string
	stmt   *sql.Stmt
	tracer *dbTracer
}

// ExecContext wraps sql.Stmt.ExecContext with tracing.
func (ts *TracedStmt) ExecContext(ctx context.Context, args ...any) (sql.Result, error) {
	var (
		err    error
		result sql.Result
	)
	ts.tracer.Trace(ctx, ts.query, func() error {
		result, err = ts.stmt.ExecContext(ctx, args...)
		return err
	})
	return result, err
}

// QueryContext wraps sql.Stmt.QueryContext with tracing.
func (ts *TracedStmt) QueryContext(ctx context.Context, args ...any) (*sql.Rows, error) {
	var (
		err  error
		rows *sql.Rows
	)
	ts.tracer.Trace(ctx, ts.query, func() error {
		rows, err = ts.stmt.QueryContext(ctx, args...)
		return err
	})
	return rows, err
}

// QueryRowContext wraps sql.Stmt.QueryRowContext with tracing.
func (ts *TracedStmt) QueryRowContext(ctx context.Context, args ...any) *sql.Row {
	var row *sql.Row
	ts.tracer.Trace(ctx, ts.query, func() error {
		row = ts.stmt.QueryRowContext(ctx, args...)
		return nil
	})
	return row
}

// Close delegates to the underlying sql.Stmt.Close.
func (ts *TracedStmt) Close() error {
	return ts.stmt.Close()
}

// TracedTx wraps *sql.Tx with tracing.
type TracedTx struct {
	tx     *sql.Tx
	tracer *dbTracer
}

// ExecContext wraps sql.Tx.ExecContext with tracing.
func (ttx *TracedTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	var (
		err    error
		result sql.Result
	)
	ttx.tracer.Trace(ctx, query, func() error {
		result, err = ttx.tx.ExecContext(ctx, query, args...)
		return err
	})
	return result, err
}

// QueryContext wraps sql.Tx.QueryContext with tracing.
func (ttx *TracedTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	var (
		err  error
		rows *sql.Rows
	)
	ttx.tracer.Trace(ctx, query, func() error {
		rows, err = ttx.tx.QueryContext(ctx, query, args...)
		return err
	})
	return rows, err
}

// QueryRowContext wraps sql.Tx.QueryRowContext with tracing.
func (ttx *TracedTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	var row *sql.Row
	ttx.tracer.Trace(ctx, query, func() error {
		row = ttx.tx.QueryRowContext(ctx, query, args...)
		return nil
	})
	return row
}

// PrepareContext wraps sql.Tx.PrepareContext with tracing and returns a TracedStmt.
func (ttx *TracedTx) PrepareContext(ctx context.Context, query string) (*TracedStmt, error) {
	var (
		err  error
		stmt *sql.Stmt
	)
	ttx.tracer.Trace(ctx, query, func() error {
		stmt, err = ttx.tx.PrepareContext(ctx, query)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &TracedStmt{stmt: stmt, tracer: ttx.tracer, query: query}, nil
}

// Commit wraps sql.Tx.Commit with tracing.
func (ttx *TracedTx) Commit() error {
	return ttx.tracer.Trace(context.Background(), "COMMIT", func() error {
		return ttx.tx.Commit()
	})
}

// Rollback wraps sql.Tx.Rollback with tracing.
func (ttx *TracedTx) Rollback() error {
	return ttx.tracer.Trace(context.Background(), "ROLLBACK", func() error {
		return ttx.tx.Rollback()
	})
}
