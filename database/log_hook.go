package database

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"
	log "github.com/integration-system/isp-log"
	"github.com/sirupsen/logrus"
)

const (
	debugQueryCode = -1
)

type logQueryHook struct {
	level logrus.Level
}

func (hook logQueryHook) BeforeQuery(ctx context.Context, _ *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (hook logQueryHook) AfterQuery(_ context.Context, q *pg.QueryEvent) error {
	m := log.WithMetadata(map[string]interface{}{
		"elapsed_time": time.Since(q.StartTime),
	})
	if query, err := q.FormattedQuery(); err != nil {
		m.Log(hook.level, debugQueryCode, err)
	} else {
		m.Log(hook.level, debugQueryCode, query)
	}
	return nil
}

func NewLogQueryHook(level logrus.Level) pg.QueryHook {
	return logQueryHook{
		level: level,
	}
}
