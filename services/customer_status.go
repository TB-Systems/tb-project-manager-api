package services

import (
	"context"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
	"github.com/google/uuid"
)

type CustomerStatusSync interface {
	Sync(ctx context.Context, customerID uuid.UUID) errors.ApiError
	SyncMany(ctx context.Context, customerIDs []uuid.UUID) errors.ApiError
}

type customerStatusSync struct {
	repository repositories.CustomerStatus
	now        func() time.Time
}

func NewCustomerStatusSyncService(repository repositories.CustomerStatus) CustomerStatusSync {
	return customerStatusSync{repository: repository, now: time.Now}
}

func (c customerStatusSync) Sync(ctx context.Context, customerID uuid.UUID) errors.ApiError {
	customerProjects, err := c.repository.ListCustomerProjects(ctx, customerID)
	if err != nil {
		return internalCustomerError("LIST_CUSTOMER_PROJECTS_FOR_STATUS_FAILED")
	}

	status := ResolveCustomerStatus(customerProjects, c.now().UTC())
	if err := c.repository.UpdateCustomerStatus(ctx, customerID, status); err != nil {
		return internalCustomerError("UPDATE_CUSTOMER_STATUS_FAILED")
	}

	return nil
}

func (c customerStatusSync) SyncMany(ctx context.Context, customerIDs []uuid.UUID) errors.ApiError {
	seen := make(map[uuid.UUID]bool, len(customerIDs))
	for _, customerID := range customerIDs {
		if customerID == uuid.Nil || seen[customerID] {
			continue
		}
		seen[customerID] = true

		if apiErr := c.Sync(ctx, customerID); apiErr != nil {
			return apiErr
		}
	}

	return nil
}

func ResolveCustomerStatus(customerProjects []models.CustomerProject, now time.Time) models.CustomerStatus {
	if len(customerProjects) == 0 {
		return models.CustomerStatusOnboarding
	}

	hasActiveLink := false
	today := beginningOfDay(now)

	for _, customerProject := range customerProjects {
		if customerProject.Status == models.CustomerProjectStatusActive {
			hasActiveLink = true
		}

		for _, invoice := range customerProject.Invoices {
			if invoice.Type == models.CustomerProjectInvoiceTypeSetupFirstHalf ||
				invoice.Type == models.CustomerProjectInvoiceTypeSetupSecondHalf {
				if invoice.Status != models.CustomerProjectInvoiceStatusPaid {
					return models.CustomerStatusOnboarding
				}
				continue
			}

			if invoice.Type != models.CustomerProjectInvoiceTypeSetupFirstHalf &&
				invoice.Type != models.CustomerProjectInvoiceTypeSetupSecondHalf &&
				(invoice.Status == models.CustomerProjectInvoiceStatusOverdue ||
					(invoice.Status == models.CustomerProjectInvoiceStatusOpen && invoice.DueDate.Before(today))) {
				return models.CustomerStatusLateMonthlyPayment
			}
		}
	}

	if hasActiveLink {
		return models.CustomerStatusActive
	}

	return models.CustomerStatusCanceled
}
