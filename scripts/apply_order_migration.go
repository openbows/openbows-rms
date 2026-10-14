//go:build ignore

package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	data, _ := os.ReadFile("configs/config.yaml")
	var cfg struct {
		Database struct {
			DSN string `yaml:"dsn"`
		} `yaml:"database"`
	}
	_ = yaml.Unmarshal(data, &cfg)
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil { panic(err) }
	var name *string
	_ = db.Raw("SELECT to_regclass('public.orders')").Scan(&name).Error
	if name != nil && *name != "" {
		fmt.Println("ALREADY_EXISTS")
		return
	}
	sql, _ := os.ReadFile("migrations/000004_phase2_order.up.sql")
	if err := db.Exec(string(sql)).Error; err != nil {
		panic(err)
	}
	fmt.Println("MIGRATED_OK")
}
