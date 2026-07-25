package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
	"github.com/TB-Systems/tb-project-manager-api/services"
)

type DailyBillingJob struct {
	invoiceRepository repositories.CustomerProjectInvoice
	billingSync       services.CustomerProjectBillingSync
	customerStatus    services.CustomerStatusSync
	now               func() time.Time
}

func NewDailyBillingJob(
	invoiceRepository repositories.CustomerProjectInvoice,
	billingSync services.CustomerProjectBillingSync,
	customerStatus services.CustomerStatusSync,
) *DailyBillingJob {
	return &DailyBillingJob{
		invoiceRepository: invoiceRepository,
		billingSync:       billingSync,
		customerStatus:    customerStatus,
		now:               time.Now,
	}
}

func (d *DailyBillingJob) Run(ctx context.Context) error {
	now := d.now()
	referenceMonth := firstDayOfMonth(now)

	invoices, err := d.monthlyInvoices(ctx, referenceMonth)
	if err != nil {
		return err
	}

	customerIDs, err := d.invoiceRepository.CreateMonthlyInvoices(ctx, invoices)
	if err != nil {
		return fmt.Errorf("create monthly invoices: %w", err)
	}

	if err := apiErrorToError(d.customerStatus.SyncMany(ctx, customerIDs)); err != nil {
		return err
	}

	if err := apiErrorToError(d.billingSync.SyncOverdue(ctx)); err != nil {
		return err
	}

	return nil
}

func (d *DailyBillingJob) monthlyInvoices(ctx context.Context, referenceMonth time.Time) ([]models.CustomerProjectInvoice, error) {
	customerProjects, err := d.invoiceRepository.ListMonthlyInvoiceCandidates(ctx)
	if err != nil {
		return nil, fmt.Errorf("list monthly invoice candidates: %w", err)
	}

	invoices := make([]models.CustomerProjectInvoice, 0, len(customerProjects))
	for _, customerProject := range customerProjects {
		term, ok := activeTerm(customerProject.Terms)
		if !ok || term.MonthlyValue <= 0 {
			continue
		}

		if customerProject.Project.Status != models.ProjectStatusLive {
			continue
		}

		if !setupPaid(customerProject.Invoices) || monthlyInvoiceExists(customerProject.Invoices, referenceMonth) {
			continue
		}

		invoices = append(invoices, models.CustomerProjectInvoice{
			CustomerProjectID: customerProject.ID,
			Type:              models.CustomerProjectInvoiceTypeMonthly,
			ReferenceMonth:    &referenceMonth,
			Amount:            term.MonthlyValue,
			DueDate:           dueDate(referenceMonth, term.DueDay),
			Status:            models.CustomerProjectInvoiceStatusOpen,
		})
	}

	return invoices, nil
}

func activeTerm(terms []models.CustomerProjectTerm) (models.CustomerProjectTerm, bool) {
	for _, term := range terms {
		if term.Active {
			return term, true
		}
	}

	return models.CustomerProjectTerm{}, false
}

func setupPaid(invoices []models.CustomerProjectInvoice) bool {
	for _, invoice := range invoices {
		switch invoice.Type {
		case models.CustomerProjectInvoiceTypeSetupFirstHalf, models.CustomerProjectInvoiceTypeSetupSecondHalf:
			if invoice.Status != models.CustomerProjectInvoiceStatusPaid {
				return false
			}
		}
	}

	return true
}

func monthlyInvoiceExists(invoices []models.CustomerProjectInvoice, referenceMonth time.Time) bool {
	for _, invoice := range invoices {
		if invoice.Type != models.CustomerProjectInvoiceTypeMonthly || invoice.ReferenceMonth == nil {
			continue
		}

		if sameMonth(*invoice.ReferenceMonth, referenceMonth) {
			return true
		}
	}

	return false
}

func firstDayOfMonth(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, value.Location())
}

func dueDate(referenceMonth time.Time, dueDay int) time.Time {
	lastDay := time.Date(referenceMonth.Year(), referenceMonth.Month()+1, 0, 0, 0, 0, 0, referenceMonth.Location()).Day()
	if dueDay > lastDay {
		dueDay = lastDay
	}
	if dueDay < 1 {
		dueDay = 1
	}

	return time.Date(referenceMonth.Year(), referenceMonth.Month(), dueDay, 0, 0, 0, 0, referenceMonth.Location())
}

func sameMonth(left time.Time, right time.Time) bool {
	return left.Year() == right.Year() && left.Month() == right.Month()
}

func apiErrorToError(apiErr errors.ApiError) error {
	if apiErr == nil {
		return nil
	}

	return fmt.Errorf("api error status %d", apiErr.GetStatus())
}
