package postgres

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jp255ft-debug/payment-core/internal/core/domain"
)

// paymentDB Ã© a representaÃ§Ã£o flat do pagamento no banco de dados.
type paymentDB struct {
	ID               string
	ExternalID       *string
	Status           string
	PaymentType      string
	Method           string
	AmountCents      int64
	FeeCents         int64
	NetAmountCents   int64
	PayerName        string
	PayerTaxID       string
	PayerEmail       string
	PayerPixKey      *string
	PayerPixType     *string
	PayeeName        string
	PayeeTaxID       string
	PayeePixKey      *string
	PayeePixType     *string
	PayeeBankCode    *string
	PayeeBranch      *string
	PayeeAccount     *string
	PayeeAccountType *string
	Description      string
	ExpiresAt        *time.Time
	ProcessedAt      *time.Time
	SettledAt        *time.Time
	FailedAt         *time.Time
	CancelledAt      *time.Time
	FailureReason    *string
	Metadata         map[string]string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Version          int
}

// fromDomain converte um domain.Payment para paymentDB.
func fromDomain(p *domain.Payment) *paymentDB {
	return &paymentDB{
		ID:               p.ID(),
		ExternalID:       strPtr(p.ExternalID()),
		Status:           string(p.Status()),
		PaymentType:      string(p.PaymentType()),
		Method:           string(p.Method()),
		AmountCents:      p.Amount().Cents(),
		FeeCents:         p.Fee().Cents(),
		NetAmountCents:   p.NetAmount().Cents(),
		PayerName:        p.Payer().Name,
		PayerTaxID:       p.Payer().TaxID,
		PayerEmail:       p.Payer().Email,
		PayerPixKey:      strPtr(p.Payer().PixKey.Value()),
		PayerPixType:     strPtr(string(p.Payer().PixKey.Type())),
		PayeeName:        p.Payee().Name,
		PayeeTaxID:       p.Payee().TaxID,
		PayeePixKey:      strPtr(p.Payee().PixKey.Value()),
		PayeePixType:     strPtr(string(p.Payee().PixKey.Type())),
		PayeeBankCode:    strPtrSafe(p.Payee().BankAccount, func(ba *domain.BankAccount) string { return ba.BankCode }),
		PayeeBranch:      strPtrSafe(p.Payee().BankAccount, func(ba *domain.BankAccount) string { return ba.Branch }),
		PayeeAccount:     strPtrSafe(p.Payee().BankAccount, func(ba *domain.BankAccount) string { return ba.AccountNumber }),
		PayeeAccountType: strPtrSafe(p.Payee().BankAccount, func(ba *domain.BankAccount) string { return ba.AccountType }),
		Description:      p.Description(),
		ExpiresAt:        p.ExpiresAt(),
		ProcessedAt:      p.ProcessedAt(),
		SettledAt:        p.SettledAt(),
		FailedAt:         p.FailedAt(),
		CancelledAt:      p.CancelledAt(),
		FailureReason:    strPtr(p.FailureReason()),
		Metadata:         p.Metadata(),
		CreatedAt:        p.CreatedAt(),
		UpdatedAt:        p.UpdatedAt(),
		Version:          p.Version(),
	}
}

// toDomain converte um paymentDB para domain.Payment.
// ReconstrÃ³i o aggregate a partir dos dados do banco.
func (db *paymentDB) toDomain() (*domain.Payment, error) {
	// Reconstruir value objects
	amount, err := domain.NewMoney(db.AmountCents)
	if err != nil {
		return nil, err
	}

	payerPixKey := domain.PixKey{}
	if db.PayerPixKey != nil {
		pk, err := domain.NewPixKey(*db.PayerPixKey)
		if err != nil {
			return nil, err
		}
		payerPixKey = pk
	}

	payeePixKey := domain.PixKey{}
	if db.PayeePixKey != nil {
		pk, err := domain.NewPixKey(*db.PayeePixKey)
		if err != nil {
			return nil, err
		}
		payeePixKey = pk
	}

	payer := domain.Payer{
		Name:   db.PayerName,
		TaxID:  db.PayerTaxID,
		Email:  db.PayerEmail,
		PixKey: payerPixKey,
	}

	var bankAccount *domain.BankAccount
	if db.PayeeBankCode != nil {
		bankAccount = &domain.BankAccount{
			BankCode:      *db.PayeeBankCode,
			Branch:        *db.PayeeBranch,
			AccountNumber: *db.PayeeAccount,
			AccountType:   *db.PayeeAccountType,
		}
	}

	payee := domain.Payee{
		Name:        db.PayeeName,
		TaxID:       db.PayeeTaxID,
		PixKey:      payeePixKey,
		BankAccount: bankAccount,
	}

	// Criar o aggregate usando NewPayment (reconstituiÃ§Ã£o)
	payment, err := domain.NewPayment(
		domain.PaymentType(db.PaymentType),
		domain.PaymentMethod(db.Method),
		amount,
		payer,
		payee,
		db.Description,
		db.Metadata,
	)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

// scanPayment lÃª uma linha do pgx.Row (QueryRow ou Rows.Next).
func scanPayment(row pgx.Row) (*paymentDB, error) {
	var db paymentDB
	err := row.Scan(
		&db.ID, &db.ExternalID, &db.Status, &db.PaymentType, &db.Method,
		&db.AmountCents, &db.FeeCents, &db.NetAmountCents,
		&db.PayerName, &db.PayerTaxID, &db.PayerEmail, &db.PayerPixKey, &db.PayerPixType,
		&db.PayeeName, &db.PayeeTaxID, &db.PayeePixKey, &db.PayeePixType,
		&db.PayeeBankCode, &db.PayeeBranch, &db.PayeeAccount, &db.PayeeAccountType,
		&db.Description, &db.ExpiresAt, &db.ProcessedAt, &db.SettledAt, &db.FailedAt, &db.CancelledAt,
		&db.FailureReason, &db.Metadata, &db.CreatedAt, &db.UpdatedAt, &db.Version,
	)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

// strPtr converte string para *string (nil se vazia).
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// strPtrSafe converte um campo aninhado opcional para *string.
func strPtrSafe[T any](ptr *T, getter func(T) string) *string {
	if ptr == nil {
		return nil
	}
	s := getter(*ptr)
	if s == "" {
		return nil
	}
	return &s
}
