package domain

import "time"

// 负责业务
type User struct {
	Id       int64
	Email    string
	Password string
	Ctime    time.Time
}
