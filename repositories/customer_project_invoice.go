package repositories

import (
	"context"
	"time"

	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CustomerProjectInvoice interface {
	FindByID(ctx context.Context, id uuid.UUID) (models.CustomerProjectInvoice, error)
	Create(ctx context.Context, invoice models.CustomerProjectInvoice) (models.CustomerProjectInvoice, error)
	Update(ctx context.Context, invoice models.CustomerProjectInvoice) (models.CustomerProjectInvoice, error)
	Pay(ctx context.Context, id uuid.UUID, paidAt time.Time) (models.CustomerProjectInvoice, error)
	Unpay(ctx context.Context, id uuid.UUID, status models.CustomerProjectInvoiceStatus) (models.CustomerProjectInvoice, error)
	MarkOverdue(ctx context.Context, today time.Time) ([]uuid.UUID, error)
	ListMonthlyInvoiceCandidates(ctx context.Context) ([]models.CustomerProject, error)
	CreateMonthlyInvoices(ctx context.Context, invoices []models.CustomerProjectInvoice) ([]uuid.UUID, error)
}

type customerProjectInvoice struct {
	db *gorm.DB
}

func NewCustomerProjectInvoiceRepository(db *gorm.DB) CustomerProjectInvoice {
	return customerProjectInvoice{db: db}
}

func (c customerProjectInvoice) FindByID(ctx context.Context, id uuid.UUID) (models.CustomerProjectInvoice, error) {
	var invoice models.CustomerProjectInvoice
	if err := c.db.WithContext(ctx).
		Preload("CustomerProject").
		First(&invoice, "id = ?", id).Error; err != nil {
		return models.CustomerProjectInvoice{}, err
	}

	return invoice, nil
}

func (c customerProjectInvoice) Create(ctx context.Context, invoice models.CustomerProjectInvoice) (models.CustomerProjectInvoice, error) {
	if err := c.db.WithContext(ctx).Create(&invoice).Error; err != nil {
		return models.CustomerProjectInvoice{}, err
	}

	return c.FindByID(ctx, invoice.ID)
}

func (c customerProjectInvoice) Update(ctx context.Context, invoice models.CustomerProjectInvoice) (models.CustomerProjectInvoice, error) {
	err := c.db.WithContext(ctx).
		Model(&models.CustomerProjectInvoice{}).
		Where("id = ?", invoice.ID).
		Updates(map[string]interface{}{
			"type":            invoice.Type,
			"reference_month": invoice.ReferenceMonth,
			"amount":          invoice.Amount,
			"due_date":        invoice.DueDate,
			"status":          invoice.Status,
			"paid_at":         invoice.PaidAt,
		}).
		Error
	if err != nil {
		return models.CustomerProjectInvoice{}, err
	}

	return c.FindByID(ctx, invoice.ID)
}

func (c customerProjectInvoice) Pay(ctx context.Context, id uuid.UUID, paidAt time.Time) (models.CustomerProjectInvoice, error) {
	err := c.db.WithContext(ctx).
		Model(&models.CustomerProjectInvoice{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  models.CustomerProjectInvoiceStatusPaid,
			"paid_at": paidAt,
		}).
		Error
	if err != nil {
		return models.CustomerProjectInvoice{}, err
	}

	return c.FindByID(ctx, id)
}

func (c customerProjectInvoice) Unpay(ctx context.Context, id uuid.UUID, status models.CustomerProjectInvoiceStatus) (models.CustomerProjectInvoice, error) {
	err := c.db.WithContext(ctx).
		Model(&models.CustomerProjectInvoice{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  status,
			"paid_at": nil,
		}).
		Error
	if err != nil {
		return models.CustomerProjectInvoice{}, err
	}

	return c.FindByID(ctx, id)
}

func (c customerProjectInvoice) MarkOverdue(ctx context.Context, today time.Time) ([]uuid.UUID, error) {
	var customerIDs []uuid.UUID
	err := c.db.WithContext(ctx).
		Model(&models.CustomerProjectInvoice{}).
		Distinct("customer_projects.customer_id").
		Joins("JOIN customer_projects ON customer_projects.id = customer_project_invoices.customer_project_id").
		Where("customer_project_invoices.status = ? AND customer_project_invoices.due_date < ?", models.CustomerProjectInvoiceStatusOpen, today).
		Pluck("customer_projects.customer_id", &customerIDs).
		Error
	if err != nil {
		return nil, err
	}

	err = c.db.WithContext(ctx).
		Model(&models.CustomerProjectInvoice{}).
		Where("status = ? AND due_date < ?", models.CustomerProjectInvoiceStatusOpen, today).
		Update("status", models.CustomerProjectInvoiceStatusOverdue).
		Error
	if err != nil {
		return nil, err
	}

	return customerIDs, nil
}

func (c customerProjectInvoice) ListMonthlyInvoiceCandidates(ctx context.Context) ([]models.CustomerProject, error) {
	var customerProjects []models.CustomerProject
	err := c.db.WithContext(ctx).
		Preload("Project").
		Preload("Terms", "active = ?", true).
		Preload("Invoices").
		Where("status = ?", models.CustomerProjectStatusActive).
		Find(&customerProjects).
		Error
	if err != nil {
		return nil, err
	}

	return customerProjects, nil
}

func (c customerProjectInvoice) CreateMonthlyInvoices(ctx context.Context, invoices []models.CustomerProjectInvoice) ([]uuid.UUID, error) {
	if len(invoices) == 0 {
		return nil, nil
	}

	err := c.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "customer_project_id"},
				{Name: "type"},
				{Name: "reference_month"},
			},
			DoNothing: true,
		}).
		Create(&invoices).
		Error
	if err != nil {
		return nil, err
	}

	customerProjectIDs := make([]uuid.UUID, 0, len(invoices))
	for _, invoice := range invoices {
		customerProjectIDs = append(customerProjectIDs, invoice.CustomerProjectID)
	}

	var customerIDs []uuid.UUID
	err = c.db.WithContext(ctx).
		Model(&models.CustomerProject{}).
		Distinct("customer_id").
		Where("id IN ?", customerProjectIDs).
		Pluck("customer_id", &customerIDs).
		Error
	if err != nil {
		return nil, err
	}

	return customerIDs, nil
}
