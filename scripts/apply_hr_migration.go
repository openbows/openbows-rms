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
	data, _ := os.ReadFile("configs/config.yaml")
	var cfg struct {
		Database struct {
			DSN string `yaml:"dsn"`
		} `yaml:"database"`
	}
	_ = yaml.Unmarshal(data, &cfg)
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		panic(err)
	}
	var name *string
	_ = db.Raw("SELECT to_regclass('public.employees')").Scan(&name).Error
	if name != nil && *name != "" {
		fmt.Println("ALREADY_EXISTS")
		return
	}
	sqlBytes, _ := os.ReadFile("migrations/000014_phase6_hr.up.sql")
	sql := string(sqlBytes)
	// 按语句拆分执行，避免多语句 Exec 限制
	parts := splitSQL(sql)
	for _, p := range parts {
		if err := db.Exec(p).Error; err != nil {
			panic(fmt.Errorf("%w\nSQL: %s", err, p[:min(120, len(p))]))
		}
	}
	fmt.Println("MIGRATED_OK")
}

func splitSQL(sql string) []string {
	var parts []string
	var b strings.Builder
	for _, line := range strings.Split(sql, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "--") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
		if strings.HasSuffix(trim, ";") {
			stmt := strings.TrimSpace(b.String())
			if stmt != "" {
				parts = append(parts, stmt)
			}
			b.Reset()
		}
	}
	if rest := strings.TrimSpace(b.String()); rest != "" {
		parts = append(parts, rest)
	}
	return parts
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
