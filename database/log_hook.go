package database

import (
	"github.com/go-pg/pg"
	log "github.com/integration-system/isp-log"
	"github.com/sirupsen/logrus"
)

const (
	debugQueryCode = -1
)

type logQueryHook struct {
	level logrus.Level
}

func (hook logQueryHook) BeforeQuery(q *pg.QueryEvent) {
}

func (hook logQueryHook) AfterQuery(q *pg.QueryEvent) {
	if query, err := q.FormattedQuery(); err != nil {
		log.Log(hook.level, debugQueryCode, err)
	} else {
		log.Log(hook.level, debugQueryCode, query)
	}
}

func NewLogQueryHook(level logrus.Level) pg.QueryHook {
	return logQueryHook{
		level: level,
	}
}
