package ports

import "github.com/hsulipe/lennie/login/internal/domain"

type UserRepository interface {
	FindIdentityByProvider(provider, providerID string) (*domain.UserIdentity, error)
	FindUserByID(id string) (*domain.User, error)
	IdentityExists(provider, providerID string) (bool, error)
	CreateUserWithIdentity(user *domain.User, identity *domain.UserIdentity) error
}
