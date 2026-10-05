package repository

import (
	"context"
	"github.com/jason127vip-dot/Go-Sales/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
	"testing"
)

func TestBranchScope(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost user=test dbname=test", PreferSimpleProtocol: true}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := model.WithBranch(context.Background(), 7)
	for _, table := range []string{"sales_orders", "sales_outbounds", "sales_invoices", "payments"} {
		t.Run(table, func(t *testing.T) {
			query := db.Table(table).Scopes(branchScope(ctx, table)).Find(&[]map[string]any{})
			sql := query.Statement.SQL.String()
			if !strings.Contains(sql, "branch_id = $1") || len(query.Statement.Vars) != 1 || query.Statement.Vars[0] != uint(7) {
				t.Fatalf("unscoped query: %s %v", sql, query.Statement.Vars)
			}
			if table != "sales_orders" && !strings.Contains(sql, table+".sales_order_id IN (SELECT id FROM sales_orders") {
				t.Fatalf("missing order inheritance: %s", sql)
			}
		})
	}
}
