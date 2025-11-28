package models

import "time"

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type Expense struct {
	ID     int64   `json:"id"`
	UserID int64   `json:"-"`
	Name   string  `json:"name"`
	Tag    string  `json:"tag"`
	Amount float64 `json:"amount"`
	Date   string  `json:"date"`
}

type MonthlyExpense struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"-"`
	Name          string     `json:"name"`
	Tag           string     `json:"tag"`
	Amount        float64    `json:"amount"`
	LastAppliedAt *time.Time `json:"lastAppliedAt,omitempty"`
	LastExpenseID *int64     `json:"lastExpenseId,omitempty"`
}
