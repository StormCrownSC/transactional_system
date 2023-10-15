package databases

import (
	"Service/internal/structures"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
)

func TestCreateInvoice(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Создаем запрос и ожидание для мока
	request := structures.TransactionRequest{
		Account:  "123456",
		Currency: "USD",
		Amount:   100.0,
	}

	mock.ExpectExec("SELECT create_invoice()").
		WithArgs(request.Account, request.Currency, request.Amount).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Вызываем функцию CreateInvoice с моком базы данных
	err = CreateInvoice(db, request)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Проверяем, что все ожидаемые вызовы были выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %s", err)
	}
}

func TestWithdrawFunds(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Создаем запрос и ожидание для мока
	request := structures.TransactionRequest{
		Account:  "123456",
		Currency: "USD",
		Amount:   100.0,
	}

	mock.ExpectExec("SELECT withdraw_funds()").
		WithArgs(request.Account, request.Currency, request.Amount).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Вызываем функцию WithdrawFunds с моком базы данных
	err = WithdrawFunds(db, request)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Проверяем, что все ожидаемые вызовы были выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %s", err)
	}
}
