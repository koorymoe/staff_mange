#!/bin/bash
# ═══ فحص الحجوزات القديمة المُرجَعة — ما يكتب ولا حرف ═══
#
# ترحيل 0250_legacy_pending_reset رجّع كل حجز قديم كان يعتبر «مثبَّت»
# قبل بوابة «تواصل مع الزبون» الجديدة إلى PENDING من جديد، لأن محد
# يضمن أحد فعلاً حچى وية زبونه بالطريقة الجديدة.
#
# هذا السكربت يوريك بالضبط شنو رجع، وكم واحد لسا قاعد بلا حركة بعد
# الرجوع (يعني ما أحد مرّره من جديد لحد هسه).
#
# الاستخدام:  ./check-legacy-pending.sh
set -e
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

echo "==> 🔍 فحص فقط — ماكو أي كتابة بقاعدة البيانات"

docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange <<'SQL'
-- كم حجز انرجّع بالمجموع
SELECT count(*) AS "إجمالي الحجوزات المُرجَعة" FROM "LegacyPendingReset";

-- كم منهن لسا قاعد PENDING (محد مرّره من جديد لحد هسه)
SELECT count(*) AS "لسا PENDING بلا حركة"
FROM "LegacyPendingReset" r
JOIN "Booking" b ON b.id = r."bookingId"
WHERE b.status = 'PENDING' AND b."confirmedAt" IS NULL;

-- كم منهن اتحرّك (موظف مرّره يدوياً بعد الترحيل)
SELECT count(*) AS "اتحرّك بعد الترحيل"
FROM "LegacyPendingReset" r
JOIN "Booking" b ON b.id = r."bookingId"
WHERE b.status <> 'PENDING' OR b."confirmedAt" IS NOT NULL;

-- عيّنة من العالقين — الاسم والهاتف وتاريخ التثبيت القديم
SELECT b.code, c.name AS "الزبون", c.phone AS "الهاتف",
       r."oldConfirmedAt" AS "تاريخ التثبيت القديم", r."resetAt" AS "تاريخ الترحيل"
FROM "LegacyPendingReset" r
JOIN "Booking" b ON b.id = r."bookingId"
JOIN "Customer" c ON c.id = b."customerId"
WHERE b.status = 'PENDING' AND b."confirmedAt" IS NULL
ORDER BY r."oldConfirmedAt" ASC
LIMIT 30;
SQL
