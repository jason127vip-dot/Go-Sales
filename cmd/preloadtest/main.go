package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/jason127vip-dot/Go-Sales/config"
	"github.com/jason127vip-dot/Go-Sales/model"
)

func main() {
	cfg := config.LoadConfig()
	db, err := config.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	var order model.SalesOrder
	err = db.
		Debug().
		Preload("Branch").
		Preload("Customer").
		Preload("Lines").
		Order("sales_orders.id DESC").
		First(&order).
		Error
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(order); err != nil {
		log.Fatal(err)
	}
}
