// أداة استيراد لمرة واحدة: تنقل قاعدة المنتجات القديمة (اللي كانت Google Sheets) إلى جدول Product.
// الاستخدام: go run ./cmd/importproducts /path/to/products.json
// ملف الـJSON مصفوفة من {name, unit, defaultPrice, imageBase64}. الاستيراد آمن للتكرار:
// أي منتج اسمه موجود بالفعل بجدول Product يتم تخطيه بدل ما يتكرر.
package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"

	"staffmange-api/internal/config"
	"staffmange-api/internal/database"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type importRow struct {
	Name         string   `json:"name"`
	Unit         *string  `json:"unit"`
	DefaultPrice *float64 `json:"defaultPrice"`
	ImageBase64  *string  `json:"imageBase64"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("الاستخدام: go run ./cmd/importproducts /path/to/products.json")
	}

	_ = godotenv.Load()
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("تعذر قراءة الملف: %v", err)
	}

	var rows []importRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		log.Fatalf("تعذر تحليل الملف: %v", err)
	}

	repo := repository.NewProductRepository(db)
	existing, err := repo.List()
	if err != nil {
		log.Fatalf("تعذر جلب المنتجات الحالية: %v", err)
	}
	existingNames := make(map[string]bool, len(existing))
	for _, p := range existing {
		existingNames[p.Name] = true
	}

	created, skipped, failed := 0, 0, 0
	for _, row := range rows {
		if existingNames[row.Name] {
			skipped++
			continue
		}
		// التوفر والتصنيف يتركون افتراضي — المستورد ما يعرفهم،
		// والموظف يحددهم من إدارة المنتجات.
		if _, err := repo.Create(model.CreateProductRequest{
			Name:         row.Name,
			Unit:         row.Unit,
			DefaultPrice: row.DefaultPrice,
			ImageBase64:  row.ImageBase64,
		}); err != nil {
			log.Printf("فشل استيراد %q: %v", row.Name, err)
			failed++
			continue
		}
		existingNames[row.Name] = true
		created++
	}

	log.Printf("تم: %d منتج جديد، %d متخطى (موجود مسبقاً)، %d فشل. الإجمالي بالملف: %d", created, skipped, failed, len(rows))
}
