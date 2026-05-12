package services

import (
	"errors"

	"github.com/hsulipe/lennie/login/internal/domain"
	"github.com/hsulipe/lennie/login/internal/ports"
)

type OIDCSignInInput struct {
	Subject       string
	Email         string
	Name          string
	EmailVerified bool
}

type OIDCService struct {
	repo ports.UserRepository
}

func NewOIDCService(repo ports.UserRepository) *OIDCService {
	return &OIDCService{repo: repo}
}

// SignIn finds or provisions a user from an OIDC token's claims.
// On first login the user is created; on subsequent logins the existing
// identity is matched by (provider=google, sub).
func (s *OIDCService) SignIn(input OIDCSignInInput) (*domain.User, error) {
	identity, err := s.repo.FindIdentityByProvider("google", input.Subject)
	if err == nil {
		return s.repo.FindUserByID(identity.UserID)
	}

	if _, err := s.repo.FindUserByEmail(input.Email); err == nil {
		return nil, errors.New("email already registered with another provider")
	}

	user := &domain.User{
		ID:            newUUID(),
		Name:          input.Name,
		Email:         input.Email,
		EmailVerified: input.EmailVerified,
	}
	newIdentity := &domain.UserIdentity{
		ID:         newUUID(),
		UserID:     user.ID,
		Provider:   "google",
		ProviderID: input.Subject,
	}
	if err := s.repo.CreateUserWithIdentity(user, newIdentity); err != nil {
		return nil, err
	}
	return user, nil
}
