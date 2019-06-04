package database

import (
	"github.com/go-pg/pg"
	"github.com/integration-system/isp-lib/logger"
	"github.com/sirupsen/logrus"
)

type logQueryHook struct {
	level logrus.Level
}

func (hook logQueryHook) BeforeQuery(q *pg.QueryEvent) {
}

func (hook logQueryHook) AfterQuery(q *pg.QueryEvent) {
	if query, err := q.FormattedQuery(); err != nil {
		logger.Warn("could not format pg query", err)
	} else {
		logger.Log(hook.level, query)
	}
}

func NewLogQueryHook(level logrus.Level) pg.QueryHook {
	return logQueryHook{
		level: level,
	}
}
