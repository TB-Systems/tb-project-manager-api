package repositories

import (
	"context"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Customer interface {
	List(ctx context.Context, params commonsmodels.PaginatedParams) ([]models.Customer, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (models.Customer, error)
	Create(ctx context.Context, customer models.Customer) (models.Customer, error)
	Update(ctx context.Context, customer models.Customer) (models.Customer, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SlugExists(ctx context.Context, slug string, exceptID *uuid.UUID) (bool, error)
	DocumentExists(ctx context.Context, document string, exceptID *uuid.UUID) (bool, error)
	EmailExists(ctx context.Context, email string, exceptID *uuid.UUID) (bool, error)
}

type customer struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) Customer {
	return customer{db: db}
}

func (c customer) List(ctx context.Context, params commonsmodels.PaginatedParams) ([]models.Customer, int64, error) {
	var customers []models.Customer
	var total int64

	query := c.db.WithContext(ctx).Model(&models.Customer{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Find(&customers).
		Error
	if err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

func (c customer) FindByID(ctx context.Context, id uuid.UUID) (models.Customer, error) {
	var customer models.Customer
	if err := c.db.WithContext(ctx).First(&customer, "id = ?", id).Error; err != nil {
		return models.Customer{}, err
	}

	return customer, nil
}

func (c customer) Create(ctx context.Context, customer models.Customer) (models.Customer, error) {
	if err := c.db.WithContext(ctx).Create(&customer).Error; err != nil {
		return models.Customer{}, err
	}

	return customer, nil
}

func (c customer) Update(ctx context.Context, customer models.Customer) (models.Customer, error) {
	err := c.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", customer.ID).
		Updates(map[string]interface{}{
			"name":          customer.Name,
			"slug":          customer.Slug,
			"document":      customer.Document,
			"document_type": customer.DocumentType,
			"email":         customer.Email,
			"phone":         customer.Phone,
			"status":        customer.Status,
			"url":           customer.URL,
		}).
		Error
	if err != nil {
		return models.Customer{}, err
	}

	return c.FindByID(ctx, customer.ID)
}

func (c customer) Delete(ctx context.Context, id uuid.UUID) error {
	return c.db.WithContext(ctx).Delete(&models.Customer{}, "id = ?", id).Error
}

func (c customer) SlugExists(ctx context.Context, slug string, exceptID *uuid.UUID) (bool, error) {
	return c.exists(ctx, "slug = ?", slug, exceptID)
}

func (c customer) DocumentExists(ctx context.Context, document string, exceptID *uuid.UUID) (bool, error) {
	return c.exists(ctx, "document = ?", document, exceptID)
}

func (c customer) EmailExists(ctx context.Context, email string, exceptID *uuid.UUID) (bool, error) {
	return c.exists(ctx, "email = ?", email, exceptID)
}

func (c customer) exists(ctx context.Context, condition string, value string, exceptID *uuid.UUID) (bool, error) {
	var total int64

	query := c.db.WithContext(ctx).Model(&models.Customer{}).Where(condition, value)
	if exceptID != nil {
		query = query.Where("id <> ?", *exceptID)
	}

	if err := query.Count(&total).Error; err != nil {
		return false, err
	}

	return total > 0, nil
}
