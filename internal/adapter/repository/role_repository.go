package repository

import (
	"context"
	"ecommerce/user-service/internal/core/domain/entity"
	"ecommerce/user-service/internal/core/domain/model"
	"errors"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type RoleRepositoryInterface interface {
	GetAll(ctx context.Context, search string) ([]entity.RoleEntity, error)
	GetByID(ctx context.Context, id int64) (*entity.RoleEntity, error)
	Create(ctx context.Context, req entity.RoleEntity) error
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, req entity.RoleEntity) error
}

type RoleRepository struct {
	db *gorm.DB
}

// Create implements [RoleRepositoryInterface].
func (r *RoleRepository) Create(ctx context.Context, req entity.RoleEntity) error {
	modelRole := model.Role{
		Name: req.Name,
	}

	if err := r.db.WithContext(ctx).Create(&modelRole).Error; err != nil {
		log.Errorf("[RoleRepository-1] Create: %v", err)
		return err
	}

	return nil
}

// Delete implements [RoleRepositoryInterface].
func (r *RoleRepository) Delete(ctx context.Context, id int64) error {
	modelRole := model.Role{}

	if err := r.db.WithContext(ctx).Where("id = ?", id).Preload("Users").First(&modelRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Infof("[RoleRepository-1] Delete: Role not found")
			return err
		}
		log.Errorf("[RoleRepository-2] Delete: %v", err)
		return err
	}

	if len(modelRole.Users) > 0 {
		err := errors.New("400")
		log.Infof("[RoleRepository-3] Delete: Role is associated with users")
		return err
	}

	if err := r.db.WithContext(ctx).Delete(&modelRole).Error; err != nil {
		log.Errorf("[Rolerepository-4] Delete: %v", err)
		return err
	}

	return nil
}

// GetAll implements [RoleRepositoryInterface].
func (r *RoleRepository) GetAll(ctx context.Context, search string) ([]entity.RoleEntity, error) {
	modelRoles := []model.Role{}
	if err := r.db.WithContext(ctx).Where("name ILike ?", "%"+search+"%").Find(&modelRoles).Error; err != nil {
		return nil, err
	}

	if len(modelRoles) == 0 {
		err := errors.New("404")
		log.Infof("[RoleRepository-2] GetAll: No role found")
		return nil, err
	}

	entityRole := []entity.RoleEntity{}
	for _, modelRoles := range modelRoles {
		entityRole = append(entityRole, entity.RoleEntity{
			ID:   modelRoles.ID,
			Name: modelRoles.Name,
		})
	}

	return entityRole, nil
}

// GetByID implements [RoleRepositoryInterface].
func (r *RoleRepository) GetByID(ctx context.Context, id int64) (*entity.RoleEntity, error) {
	modelRoles := model.Role{}
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&modelRoles).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := errors.New("404")
			log.Infof("[RoleRepository-1] GetByID: Role not found")
			return nil, err
		}
		log.Errorf("[RoleRepository-2] GetAll: %v", err)
		return nil, err
	}

	return &entity.RoleEntity{
		ID:   modelRoles.ID,
		Name: modelRoles.Name,
	}, nil
}

// Update implements [RoleRepositoryInterface].
func (r *RoleRepository) Update(ctx context.Context, req entity.RoleEntity) error {
	modelRole := model.Role{}

	if err := r.db.WithContext(ctx).Where("id = ?", req.ID).First(&modelRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[RoleRepository-1] Update: Role not found")
			return err
		}
		log.Errorf("[RoleRepository-2] Update: %v", err)
		return err
	}

	modelRole.Name = req.Name
	if err := r.db.WithContext(ctx).Save(&modelRole).Error; err != nil {
		log.Errorf("[RoleRepository-3] Update: %v", err)
		return err
	}

	return nil
}

func NewRoleRepository(db *gorm.DB) RoleRepositoryInterface {
	return &RoleRepository{db: db}
}
