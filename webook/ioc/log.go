package ioc

import (
	"webook/pkg/logger"
)

func InitLogger() logger.LoggerV1 {
	return logger.NewNoOpLogger()
}
