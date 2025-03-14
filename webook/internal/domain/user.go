package domain

import "time"

// 负责业务
type User struct {
	Id       int64
	Email    string
	Password string

	Nickname string
	// YYYY-MM-DD
	Birthday time.Time
	AboutMe  string

	Phone string

	// UTC 0 的时区
	Ctime time.Time

	//Addr Address
}
