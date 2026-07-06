package postgres

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
)

// transferDB é a representação flat da transferência no banco de dados.
type transferDB struct {
	ID              string
	ExternalID      *string
	Status          string
	AmountCents     int64
	PayeeName       string
	PayeeTaxID      string
	PayeePixKey     string
	PayeePixType    *string
	PayeeBankCode   *string
	PayeeBranch     *string
	PayeeAccount    *string
	PayeeAccountType *string
	Provider        string
	Description     string
	Metadata        map[string]string
	ProcessedAt     *time.Time
	SettledAt       *time.Time
	FailedAt        *time.Time
	CancelledAt     *time.Time
	FailureReason   *string
	Version         int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// fromDomain converte um entity.Transfer para transferDB.
func fromDomain(t *entity.Transfer) *transferDB {
	payee := t.Payee()
	return &transferDB{
		ID:           t.ID(),
		ExternalID:   strPtr(t.ExternalID()),
		Status:       string(t.Status()),
		AmountCents:  t.AmountCents(),
		PayeeName:    payee.Name,
		PayeeTaxID:   payee.TaxID,
		PayeePixKey:  payee.PixKey,
		PayeePixType: strPtr(payee.PixType),
		PayeeBankCode:   strPtrFromBank(payee.Bank, func(b entity.BankAccount) string { return b.BankCode }),
		PayeeBranch:     strPtrFromBank(payee.Bank, func(b entity.BankAccount) string { return b.Branch }),
		PayeeAccount:    strPtrFromBank(payee.Bank, func(b entity.BankAccount) string { return b.Account }),
		PayeeAccountType: strPtrFromBank(payee.Bank, func(b entity.BankAccount) string { return b.AccountType }),
		Provider:      string(t.Provider()),
		Description:   t.Description(),
		Metadata:      t.Metadata(),
		ProcessedAt:   t.ProcessedAt(),
		SettledAt:     t.SettledAt(),
		FailedAt:      t.FailedAt(),
		CancelledAt:   nil, // não exposto como getter
		FailureReason: strPtr(t.FailureReason()),
		Version:       t.Version(),
		CreatedAt:     t.CreatedAt(),
		UpdatedAt:     t.UpdatedAt(),
	}
}

// toDomain reconstrói um entity.Transfer a partir do transferDB.
func (db *transferDB) toDomain() (*entity.Transfer, error) {
	payee := entity.TransferPayee{
		Name:   db.PayeeName,
		TaxID:  db.PayeeTaxID,
		PixKey: db.PayeePixKey,
		PixType: safeStr(db.PayeePixType),
	}
	if db.PayeeBankCode != nil {
		payee.Bank = &entity.BankAccount{
			BankCode:    *db.PayeeBankCode,
			Branch:      safeStr(db.PayeeBranch),
			Account:     safeStr(db.PayeeAccount),
			AccountType: safeStr(db.PayeeAccountType),
		}
	}

	t, err := entity.NewTransfer(
		db.AmountCents,
		payee,
		entity.CashOutProvider(db.Provider),
		db.Description,
		db.Metadata,
	)
	if err != nil {
		return nil, err
	}

	switch entity.TransferStatus(db.Status) {
	case entity.TransferStatusApproved:
		if err := t.Approve(safeStr(db.ExternalID)); err != nil {
			return nil, err
		}
	case entity.TransferStatusSettled:
		if err := t.Approve(safeStr(db.ExternalID)); err != nil {
			return nil, err
		}
		if err := t.Settle(); err != nil {
			return nil, err
		}
	case entity.TransferStatusFailed:
		t.Fail(safeStr(db.FailureReason))
	case entity.TransferStatusCancelled:
		t.Cancel(safeStr(db.FailureReason))
	}

	return t, nil
}

// scanTransfer lê uma linha do pgx.Row (QueryRow ou Rows.Next).
func scanTransfer(row pgx.Row) (*transferDB, error) {
	var db transferDB
	err := row.Scan(
		&db.ID, &db.ExternalID, &db.Status,
		&db.AmountCents,
		&db.PayeeName, &db.PayeeTaxID, &db.PayeePixKey, &db.PayeePixType,
		&db.PayeeBankCode, &db.PayeeBranch, &db.PayeeAccount, &db.PayeeAccountType,
		&db.Provider, &db.Description,
		&db.Metadata,
		&db.ProcessedAt, &db.SettledAt, &db.FailedAt, &db.CancelledAt,
		&db.FailureReason,
		&db.Version, &db.CreatedAt, &db.UpdatedAt,
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

// strPtrFromBank extrai opcionalmente um campo de *BankAccount.
func strPtrFromBank[T any](ptr *T, getter func(T) string) *string {
	if ptr == nil {
		return nil
	}
	s := getter(*ptr)
	if s == "" {
		return nil
	}
	return &s
}

// safeStr extrai string de *string, retornando vazio se nil.
func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}