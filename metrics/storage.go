package metrics

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// Note
// Можно сделать общий сторадж для метрик SQL.
// 1) Делаем имплементацию дефолтных методов, она ниже.
// 2) Делаем общие метрики на SQL в этой папке.
// 3) В inbox/outbox выносим мигратор в отдельную структуру. Нам нужен
// исходный коннекшен для миграции. В остальных случаях нужны только интерфейсы на
// методы.

type QueryExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type MetricsStorage struct {
	conn QueryExecer
}

func NewMetricsStorage(conn QueryExecer) *MetricsStorage {
	return &MetricsStorage{
		conn: conn,
	}
}

func (s *MetricsStorage) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	var err error

	defer func(start time.Time) {
		s.measureSQLQuery(start, query, err)
	}(time.Now())

	return s.conn.ExecContext(ctx, query, args...)
}

func (s *MetricsStorage) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	var err error

	defer func(start time.Time) {
		s.measureSQLQuery(start, query, err)
	}(time.Now())

	return s.conn.QueryContext(ctx, query, args...)
}

func (s *MetricsStorage) measureSQLQuery(start time.Time, query string, err error) {
	query = s.cleanupQuery(query)

	storageCounter.WithLabelValues(query, s.isOK(err)).Inc()

	endms := float64(time.Since(start).Milliseconds())

	storageHistogram.WithLabelValues(query, s.isOK(err)).Observe(endms)
}

// removes double spaces, tabs, line breaks etc
func (s *MetricsStorage) cleanupQuery(query string) string {
	return strings.Join(strings.Fields(query), " ")
}

func (s *MetricsStorage) isOK(err error) string {
	if err != nil {
		return "error"
	}

	return "ok"
}
