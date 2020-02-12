package database

import (
	"context"

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
	if query, err := q.FormattedQuery(); err != nil {
		log.Log(hook.level, debugQueryCode, err)
	} else {
		log.Log(hook.level, debugQueryCode, query)
	}
	return nil
}

func NewLogQueryHook(level logrus.Level) pg.QueryHook {
	return logQueryHook{
		level: level,
	}
}
