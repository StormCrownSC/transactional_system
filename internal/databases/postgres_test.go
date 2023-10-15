package databases

import (
	"Service/internal/structures"
	"errors"
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

	// Тест 1: Успешное выполнение запроса
	request := structures.TransactionRequest{
		Account:  "123456",
		Currency: "USD",
		Amount:   100.0,
	}

	mock.ExpectExec("SELECT create_invoice()").
		WithArgs(request.Account, request.Currency, request.Amount).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = CreateInvoice(db, request)
	if err != nil {
		t.Errorf("Test 1: Expected no error, but got %v", err)
	}

	// Тест 2: Обработка ошибки базы данных
	mock.ExpectExec("SELECT create_invoice()").
		WithArgs(request.Account, request.Currency, request.Amount).
		WillReturnError(sqlmock.ErrCancelled)

	err = CreateInvoice(db, request)
	if err == nil {
		t.Error("Test 2: Expected an error, but got nil")
	}

	// Тест 3: Обработка ошибки результатов SQL
	mock.ExpectExec("SELECT create_invoice()").
		WithArgs(request.Account, request.Currency, request.Amount).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("test error")))

	err = CreateInvoice(db, request)
	if err != nil {
		t.Errorf("Test 1: Expected no error, but got %v", err)
	}

	// Тест 4: Обработка ошибки парсинга аргументов
	mock.ExpectExec("SELECT create_invoice()").
		WithArgs(request.Account, request.Currency, "invalid amount").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = CreateInvoice(db, request)
	if err == nil {
		t.Error("Test 4: Expected an error, but got nil")
	}

	// Тест 5: Обработка успешного выполнения без результата
	mock.ExpectExec("SELECT create_invoice()").
		WithArgs(request.Account, request.Currency, request.Amount).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = CreateInvoice(db, request)
	if err == nil {
		t.Error("Test 4: Expected an error, but got nil")
	}
}

func TestWithdrawFunds(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Тест 1: Успешное выполнение запроса
	request := structures.TransactionRequest{
		Account:  "123456",
		Currency: "USD",
		Amount:   100.0,
	}

	mock.ExpectExec("SELECT withdraw_funds()").
		WithArgs(request.Account, request.Currency, request.Amount).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = WithdrawFunds(db, request)
	if err != nil {
		t.Errorf("Test 1: Expected no error, but got %v", err)
	}

	// Тест 2: Обработка ошибки базы данных
	mock.ExpectExec("SELECT withdraw_funds()").
		WithArgs(request.Account, request.Currency, request.Amount).
		WillReturnError(sqlmock.ErrCancelled)

	err = WithdrawFunds(db, request)
	if err == nil {
		t.Error("Test 2: Expected an error, but got nil")
	}

	// Тест 3: Обработка ошибки результатов SQL
	mock.ExpectExec("SELECT withdraw_funds()").
		WithArgs(request.Account, request.Currency, request.Amount).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("test error")))

	err = WithdrawFunds(db, request)
	if err != nil {
		t.Errorf("Test 1: Expected no error, but got %v", err)
	}

	// Тест 4: Обработка ошибки парсинга аргументов
	mock.ExpectExec("SELECT withdraw_funds()").
		WithArgs(request.Account, request.Currency, "invalid amount").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = WithdrawFunds(db, request)
	if err == nil {
		t.Error("Test 4: Expected an error, but got nil")
	}

	// Тест 5: Обработка успешного выполнения без результата
	mock.ExpectExec("SELECT withdraw_funds()").
		WithArgs(request.Account, request.Currency, request.Amount).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = WithdrawFunds(db, request)
	if err == nil {
		t.Error("Test 4: Expected an error, but got nil")
	}
}
