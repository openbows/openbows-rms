//go:build ignore

package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	data, err := os.ReadFile("configs/config.yaml")
	if err != nil {
		panic(err)
	}
	var cfg struct {
		Database struct {
			DSN string `yaml:"dsn"`
		} `yaml:"database"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		panic(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	var name *string
	if err := db.Raw("SELECT to_regclass('public.stations')").Scan(&name).Error; err != nil {
		panic(err)
	}
	if name != nil && *name != "" {
		fmt.Println("ALREADY_EXISTS")
		return
	}
	sqlBytes, err := os.ReadFile("migrations/000002_phase1_menu.up.sql")
	if err != nil {
		panic(err)
	}
	sql := string(sqlBytes)
	// execute as one script via pgx connection
	if err := db.Exec(sql).Error; err != nil {
		// fallback: split by semicolon carefully for statements
		parts := strings.Split(sql, ";")
		for _, p := range parts {
			stmt := strings.TrimSpace(p)
			if stmt == "" || strings.HasPrefix(stmt, "--") {
				continue
			}
			if err := db.Exec(stmt).Error; err != nil {
				fmt.Println("ERR:", err.Error(), "STMT:", stmt[:min(80, len(stmt))])
				os.Exit(1)
			}
		}
	}
	fmt.Println("MIGRATED_OK")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
