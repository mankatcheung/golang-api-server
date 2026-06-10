// Package model defines the data structures and domain types used across the application.
package model

type ConvertRequest struct {
	From string  `json:"from" binding:"required,min=3,max=3"`
	To   string  `json:"to" binding:"required,min=3,max=3"`
	Amt  float64 `json:"amount" binding:"required,gt=0"`
}

type ConvertResponse struct {
	From     string  `json:"from"`
	To       string  `json:"to"`
	Amount   float64 `json:"amount"`
	Rate     float64 `json:"rate"`
	Converted float64 `json:"converted"`
}

type RatesResponse struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
}
