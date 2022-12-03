package db

import (
	"context"      // for a function to know more about environment
	"database/sql" // go's standard sql library
	"fmt"          // formatted I/O
)

// Store defines all functions to execute db queries and transactions (Tx): extend Queries with more functions (sql.DB) by composition
type Store struct { 
	*Queries // queries struct do not support transaction (only one operation on one specific table), so we need to extend its functionality by embedding it inside a struct (called composition), all individual query functions provided by Queries will be available to Store
	db *sql.DB 
}

// NewStore creates a new store
func NewStore(db *sql.DB) *Store { // a pointer pointing to an Store type struct
	return &Store{ // initialize a Store (Store{}) -> generate its address (&) -> in the future return it to a pointer with Store pointer type (*Store)
		db:      db,
		Queries: New(db), // func New(db DBTX) *Queries { return &Queries{db: db} }
	}
}

// ExecTx executes a function within a database transaction?
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error { // returned error is not exported and thus starts with lower letter e
	tx, err := store.db.BeginTx(ctx, nil) // ctx is used until the transaction is committed or rolled back.
	if err != nil {
		return err
	}

	q := New(tx) // tx is a DBTX interface
	err = fn(q) // call the input function, fn is a callback function
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr) // report two errors by combining them into one single error
		}
		return err // rollback is successful
	}

	return tx.Commit() // no error happened and return error if there is error by tx.commit()
}

// TransferTxParams contains the input parameters of the transfer transaction (the function "TransferTx" below)
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"` // Add "tags" to customize JSON keys
	ToAccountID   int64 `json:"to_account_id"` // JSON encoder (from json package) can only see, and therefore only encode, exported fields in a struct.
	Amount        int64 `json:"amount"` // only the identifiers that start with a captial letter are exported from your package
}

// TransferTxResult is the result of the transfer transaction (the function "TransferTx" below)
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// func addMoney(
// 	ctx context.Context,
// 	q *Queries,
// 	accountID1 int64,
// 	amount1 int64,
// 	accountID2 int64,
// 	amount2 int64,
// ) (account1 Account, account2 Account, err error) {
// 	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
// 		ID:     accountID1,
// 		Amount: amount1,
// 	})
// 	if err != nil {
// 		return
// 	}

// 	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
// 		ID:     accountID2,
// 		Amount: amount2,
// 	})
// 	return
// }

// TransferTx performs a money transfer from one account to the other.
// It creates the transfer, add account entries, and update accounts' balance within a single database transaction:
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) { 
	var result TransferTxResult // start with an empty result

	
	err := store.execTx(ctx, func(q *Queries) error { 
		// define the callback function for performs a money transfer
		// the callback function use the variable result and arg, making it a closure
		var err error

		// 1. create a transfer record with amount = 10 // from transer.sql.go
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		// 2. create an entry for account1 with amount = -10 // from entry.sql.go
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// 3. create an entry for account2 with amount = +10
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// 4. subtract 10 from the balance of account1 // 5. add 10 to the balance of account2
		result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:      arg.FromAccountID,
			Amount:  -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:      arg.ToAccountID,
			Amount:  arg.Amount,
		})
		if err != nil {
			return err
		}

		return err
	})

	return result, err
}