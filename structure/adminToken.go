package structure

import "time"

type AdminToken struct {
	TableName string     `sql:"admin_service.tokens" json:"-"`
	Id        int64      `json:"id"`
	UserId    int64      `json:"userId"`
	Token     string     `json:"token"`
	ExpiredAt *time.Time `json:"expiredAt"`
	CreatedAt time.Time  `json:"createdAt" sql:",null"`
}
