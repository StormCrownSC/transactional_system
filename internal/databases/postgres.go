package databases

import (
	"Service/internal/structures"
	"database/sql"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

func ConnectDB() (*sql.DB, error) {
	// Open a database connection and create a connection pool
	connStr := "host=transactional.postgres port=5432 user=admin password=admin dbname=TransactionalDB sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Checking the connection to the database
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	// Setting the maximum number of active connections (Depends on the load of the service)
	db.SetMaxOpenConns(20)

	// Setting the maximum number of inactive connections (Depends on the load of the service)
	db.SetMaxIdleConns(10)
	return db, nil
}

// CreateInvoice creates a new invoice and returns an error in case of failure
func CreateInvoice(db *sql.DB, request structures.TransactionRequest) error {
	// Prepare the SQL statement to call the stored procedure
	_, err := db.Exec("SELECT create_invoice($1, $2, $3)",
		request.Account, request.Currency, request.Amount)

	return err
}

// WithdrawFunds withdraws funds from a user's account and returns an error in case of failure
func WithdrawFunds(db *sql.DB, request structures.TransactionRequest) error {
	// Prepare the SQL statement to call the stored procedure
	_, err := db.Exec("SELECT withdraw_funds($1, $2, $3)",
		request.Account, request.Currency, request.Amount)

	return err
}

// Function for getting customer balances in all currencies
func GetClientBalances(db *sql.DB, clientAccount string) ([]structures.ClientBalance, error) {
	// Prepare the SQL statement to call the stored procedure
	rows, err := db.Query("SELECT * FROM get_client_balances($1)", clientAccount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clientBalances []structures.ClientBalance
	for rows.Next() {
		var currencyCode, actualBalanceStr, frozenBalanceStr string
		err := rows.Scan(&currencyCode, &actualBalanceStr, &frozenBalanceStr)
		if err != nil {
			return nil, err
		}

		actualBalance, _ := ParseBalanceString(actualBalanceStr)
		frozenBalance, _ := ParseBalanceString(frozenBalanceStr)

		clientBalance := structures.ClientBalance{
			CurrencyCode:  currencyCode,
			ActualBalance: actualBalance,
			FrozenBalance: frozenBalance,
		}
		clientBalances = append(clientBalances, clientBalance)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clientBalances, nil
}

// ParseBalanceString converts the balance line in the format "$1,000.00" to float64.
func ParseBalanceString(balanceStr string) (float64, error) {
	// Remove "$" and "," characters from balanceStr
	balanceStr = strings.ReplaceAll(balanceStr, "$", "")
	balanceStr = strings.ReplaceAll(balanceStr, ",", "")

	// Parse the cleaned balance string to float64
	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		return 0, err
	}
	return balance, nil
}
