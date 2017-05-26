package utils

import uuid "github.com/satori/go.uuid"

func IsUUID(id string) bool {
	_, err := uuid.FromString(id)
	if err != nil {
		return false
	}
	return true
}
