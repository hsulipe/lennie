package postgres

import (
	"errors"

	"github.com/hsulipe/lennie/login/internal/domain"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

// Constructor for UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindIdentityByProvider(provider, providerID string) (*domain.UserIdentity, error) {
	var identity domain.UserIdentity
	err := r.db.Where("provider = ? AND provider_identifier = ? AND deleted = false", provider, providerID).
		First(&identity).Error
	if err != nil {
		return nil, err
	}
	return &identity, nil
}

func (r *UserRepository) FindUserByID(id string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("id = ? AND deleted = false", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) IdentityExists(provider, providerID string) (bool, error) {
	var identity domain.UserIdentity
	err := r.db.Where("provider = ? AND provider_identifier = ? AND deleted = false", provider, providerID).
		First(&identity).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func (r *UserRepository) CreateUserWithIdentity(user *domain.User, identity *domain.UserIdentity) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return tx.Create(identity).Error
	})
}
