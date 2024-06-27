package utils

import (
	"strings"

	"github.com/google/uuid"
	"github.com/rumis/multicache/logger"
)

// UUID 生成UUID
func UUID() string {
	uid, err := uuid.NewRandom()
	if err != nil {
		logger.Error("uuid.NewRandom fail", "err", err)
		return ""
	}
	return strings.ReplaceAll(uid.String(), "-", "")
}
