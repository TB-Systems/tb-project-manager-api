package dto

import (
	"testing"

	"github.com/TB-Systems/tb-project-manager-api/models"
)

func TestCustomerRequestValidateDocument(t *testing.T) {
	t.Run("accepts valid cpf", func(t *testing.T) {
		request := validCustomerRequest()
		request.Document = "529.982.247-25"
		request.DocumentType = models.CustomerDocumentTypeCPF

		if errs := request.Validate(); len(errs) > 0 {
			t.Fatalf("Expected valid CPF, got %d errors", len(errs))
		}
	})

	t.Run("rejects invalid cpf", func(t *testing.T) {
		request := validCustomerRequest()
		request.Document = "529.982.247-26"
		request.DocumentType = models.CustomerDocumentTypeCPF

		if errs := request.Validate(); len(errs) == 0 {
			t.Fatal("Expected invalid CPF error")
		}
	})

	t.Run("accepts valid cnpj", func(t *testing.T) {
		request := validCustomerRequest()
		request.Document = "04.252.011/0001-10"
		request.DocumentType = models.CustomerDocumentTypeCNPJ

		if errs := request.Validate(); len(errs) > 0 {
			t.Fatalf("Expected valid CNPJ, got %d errors", len(errs))
		}
	})

	t.Run("rejects invalid cnpj", func(t *testing.T) {
		request := validCustomerRequest()
		request.Document = "04.252.011/0001-11"
		request.DocumentType = models.CustomerDocumentTypeCNPJ

		if errs := request.Validate(); len(errs) == 0 {
			t.Fatal("Expected invalid CNPJ error")
		}
	})

	t.Run("does not validate check digits for other document type", func(t *testing.T) {
		request := validCustomerRequest()
		request.Document = "foreign-registration-abc"
		request.DocumentType = models.CustomerDocumentTypeOther

		if errs := request.Validate(); len(errs) > 0 {
			t.Fatalf("Expected other document type to accept custom document, got %d errors", len(errs))
		}
	})
}

func validCustomerRequest() CustomerRequest {
	return CustomerRequest{
		Name:         "TB Systems",
		Slug:         "tb-systems",
		Document:     "04.252.011/0001-10",
		DocumentType: models.CustomerDocumentTypeCNPJ,
		Email:        "contact@tbsystems.com.br",
		Status:       models.CustomerStatusActive,
	}
}
