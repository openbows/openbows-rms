//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`).Error

	files, _ := filepath.Glob("migrations/*.up.sql")
	sort.Strings(files)
	for _, f := range files {
		base := filepath.Base(f)
		var exists int64
		_ = db.Raw("SELECT COUNT(1) FROM schema_migrations WHERE filename = ?", base).Scan(&exists)
		if exists > 0 {
			fmt.Println("SKIP", base)
			continue
		}
		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			panic(err)
		}
		parts := splitSQL(string(sqlBytes))
		err = db.Transaction(func(tx *gorm.DB) error {
			for _, p := range parts {
				if err := tx.Exec(p).Error; err != nil {
					return fmt.Errorf("%s: %w\nSQL: %s", base, err, truncate(p, 160))
				}
			}
			return tx.Exec("INSERT INTO schema_migrations(filename) VALUES (?)", base).Error
		})
		if err != nil {
			// 兼容历史库：若表已存在导致失败，尝试标记已应用
			if strings.Contains(err.Error(), "already exists") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				_ = db.Exec("INSERT INTO schema_migrations(filename) VALUES (?) ON CONFLICT DO NOTHING", base)
				fmt.Println("MARK", base, "(already applied)")
				continue
			}
			panic(err)
		}
		fmt.Println("OK", base)
	}
	fmt.Println("MIGRATE_ALL_DONE")
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
