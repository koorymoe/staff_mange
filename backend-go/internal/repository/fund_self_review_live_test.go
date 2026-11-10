package repository

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/model"
)

// (11-07) صاحب الفلوس ما يوافق على تسويته هو — چان المحاسب يرفع تسويته ويوافقها لحاله.
func TestFundSettlementNoSelfReviewLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	repo := NewRevolvingFundRepository(db)
	acc, boss, fund := uuid.NewString(), uuid.NewString(), uuid.NewString()
	sfx := uuid.NewString()[:6]
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'محاسب', $2, 'FINANCE', 'ACTIVE'), ($3, 'مدير', $4, 'ADMIN', 'ACTIVE')`, acc, "fa_"+sfx, boss, "fb_"+sfx)
	db.MustExec(`INSERT INTO "RevolvingFund" (id, name, balance) VALUES ($1, $2, 1000000)`, fund, "صندوق فحص "+sfx)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM "RevolvingFundTxn" WHERE "fundId" = $1`, fund) //nolint:errcheck
		db.Exec(`DELETE FROM "RevolvingFund" WHERE id = $1`, fund)          //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, acc, boss)   //nolint:errcheck
	})
	if _, err := repo.Disburse(model.DisburseRequest{FundID: fund, EmployeeID: acc, Amount: 500000}, &boss); err != nil {
		t.Fatal(err)
	}
	receipt, note := "data:image/png;base64,iVBORw0KGgo=", "مواد كهرباء للموقع"
	st, err := repo.SubmitSettlement(acc, model.SettlementRequest{FundID: fund, SpentAmount: 300000, ReturnedAmount: 200000, ReceiptImage: &receipt, Notes: &note})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ReviewSettlement(st.ID, model.ReviewSettlementRequest{Approve: true}, &acc); err == nil {
		t.Fatal("المحاسب وافق على تسويته هو")
	}
	if _, err := repo.ReviewSettlement(st.ID, model.ReviewSettlementRequest{Approve: true}, &boss); err != nil {
		t.Fatalf("المدير ما گدر يوافق: %v", err)
	}
}
