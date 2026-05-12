package services

import (
	"errors"

	"github.com/hsulipe/lennie/login/internal/domain"
	"github.com/hsulipe/lennie/login/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

type SignInService struct {
	repo ports.UserRepository
}

func NewSignInService(repo ports.UserRepository) *SignInService {
	return &SignInService{repo: repo}
}

func (s *SignInService) SignIn(provider, providerID, credentials string) (*domain.User, error) {
	identity, err := s.repo.FindIdentityByProvider(provider, providerID)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if identity.CredentialsHash == nil {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*identity.CredentialsHash), []byte(credentials)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return s.repo.FindUserByID(identity.UserID)
}
