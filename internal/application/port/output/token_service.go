package output

import (
	"learning-go/internal/domain/entity"
)

type TokenServicePort interface {
	GenerateToken(user *entity.User) (string, error)
}
