package startup

import (
	"project/webook/pkg/logger"
)

func InitLogger() logger.LoggerV1 {
	return logger.NewNoOpLogger()
}
