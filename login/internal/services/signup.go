package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/hsulipe/lennie/login/internal/domain"
	"github.com/hsulipe/lennie/login/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

type SignUpInput struct {
	Name       string
	Email      string
	Password   string
	Birthdate  string
	CPF        string
	Phone      string
	Provider   string
	ProviderID string
}

type SignUpService struct {
	repo ports.UserRepository
}

func NewSignUpService(repo ports.UserRepository) *SignUpService {
	return &SignUpService{repo: repo}
}

func (s *SignUpService) SignUp(input SignUpInput) (*domain.User, error) {
	exists, err := s.repo.IdentityExists(input.Provider, input.ProviderID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	birthdate, err := time.Parse("2006-01-02", input.Birthdate)
	if err != nil {
		return nil, errors.New("invalid birthdate format, expected YYYY-MM-DD")
	}

	user := &domain.User{
		ID:        newUUID(),
		Name:      input.Name,
		Email:     input.Email,
		CPF:       input.CPF,
		Phone:     input.Phone,
		Birthdate: birthdate,
	}

	identity := &domain.UserIdentity{
		ID:              newUUID(),
		UserID:          user.ID,
		Provider:        input.Provider,
		ProviderID:      input.ProviderID,
		CredentialsHash: string(hash),
	}

	if err := s.repo.CreateUserWithIdentity(user, identity); err != nil {
		return nil, err
	}

	return user, nil
}

func newUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
