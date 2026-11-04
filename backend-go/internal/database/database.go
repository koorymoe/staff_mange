package database

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	// (ع) 10-10 «النظام صار ثكيل»: 20 اتصال جان يخلّي الطلبات توكف بالطابور
	// وقت الزحمة، و5 خاملة بس تعني فتح اتصال جديد بكل دفعة طلبات.
	db.SetMaxOpenConns(40)
	db.SetMaxIdleConns(15)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}
