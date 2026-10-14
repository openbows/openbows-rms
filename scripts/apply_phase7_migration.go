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
	_ = db.Raw("SELECT to_regclass('public.dish_materials')").Scan(&name)
	if name == nil || *name == "" {
		runFile(db, "migrations/000015_phase7_dish_bom.up.sql")
		fmt.Println("BOM_MIGRATED_OK")
	} else {
		fmt.Println("BOM_ALREADY_EXISTS")
	}

	var cnt int64
	_ = db.Raw("SELECT COUNT(1) FROM permissions WHERE code = 'order:write' AND deleted_at IS NULL").Scan(&cnt)
	if cnt == 0 {
		runFile(db, "migrations/000016_phase7_rbac_codes.up.sql")
		fmt.Println("RBAC_MIGRATED_OK")
	} else {
		// 仍确保 admin 拥有全部权限
		_ = db.Exec(`INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, p.id FROM permissions p
WHERE p.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp
      WHERE rp.role_id = 1 AND rp.permission_id = p.id AND rp.deleted_at IS NULL
  )`)
		fmt.Println("RBAC_ALREADY_EXISTS")
	}
}

func runFile(db *gorm.DB, path string) {
	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	for _, p := range splitSQL(string(sqlBytes)) {
		if err := db.Exec(p).Error; err != nil {
			panic(fmt.Errorf("%s: %w", path, err))
		}
	}
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
