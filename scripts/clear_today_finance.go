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
  data,_ := os.ReadFile("configs/config.yaml")
  var cfg struct{ Database struct{ DSN string `yaml:"dsn"` } `yaml:"database"` }
  _ = yaml.Unmarshal(data, &cfg)
  db,_ := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
  r := db.Exec("UPDATE settlements SET deleted_at=NOW() WHERE business_day=CURRENT_DATE AND deleted_at IS NULL")
  fmt.Println("SETTLEMENTS_CLEARED", r.RowsAffected)
  r2 := db.Exec("UPDATE settlement_items SET deleted_at=NOW() WHERE settlement_id IN (SELECT id FROM settlements WHERE business_day=CURRENT_DATE) AND deleted_at IS NULL")
  fmt.Println("ITEMS_CLEARED", r2.RowsAffected)
  r3 := db.Exec("UPDATE reconciliations SET deleted_at=NOW() WHERE business_day=CURRENT_DATE AND deleted_at IS NULL")
  fmt.Println("RECONS_CLEARED", r3.RowsAffected)
  r4 := db.Exec("UPDATE reconciliation_items SET deleted_at=NOW() WHERE reconciliation_id IN (SELECT id FROM reconciliations WHERE business_day=CURRENT_DATE) AND deleted_at IS NULL")
  fmt.Println("RECON_ITEMS_CLEARED", r4.RowsAffected)
}
