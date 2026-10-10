#!/bin/bash
# ═══ تمرير الحجوزات القديمة المُرجَعة — يكتب فعلاً، بعد فحص وموافقة ═══
#
# استعمل check-legacy-pending.sh أولاً وشوف العدد والأسماء قبل ما
# تشغّل هذا. هذا يفترض إن الزبون فعلاً تواصل معه أحد بالماضي (قبل
# البوابة الجديدة) ويعيده لحالة «تم التثبيت» مباشرة، متجاوزاً خطوة
# «تواصل مع الزبون» — لأنها صارت بالفعل قبل أشهر ومحد يقدر يعيدها.
#
# لو تفضّل الموظف المسؤول يمرّرهن بنفسه واحداً واحداً من الشاشة
# (أدق، بس شغل يدوي)، لا تشغّل هذا السكربت إطلاقاً.
#
# الاستخدام:  ./resume-legacy-pending.sh --confirm
set -e
DB_CONTAINER="${DB_CONTAINER:-staff_mange-db-1}"

if [ "$1" != "--confirm" ]; then
  echo "هذا السكربت يكتب على قاعدة البيانات."
  echo "شغّل check-legacy-pending.sh أولاً وتأكد من العدد والأسماء."
  echo "لو متأكد: ./resume-legacy-pending.sh --confirm"
  exit 1
fi

echo "==> يعيد تثبيت الحجوزات القديمة من قيمها المحفوظة قبل الترحيل"

docker exec -i "$DB_CONTAINER" psql -U staffmange -d staffmange <<'SQL'
BEGIN;

UPDATE "Booking" b SET
  "confirmedAt" = r."oldConfirmedAt",
  "confirmationContactedAt" = r."oldContactedAt",
  status = 'CONFIRMED',
  "updatedAt" = now()
FROM "LegacyPendingReset" r
WHERE r."bookingId" = b.id
  AND b.status = 'PENDING'
  AND b."confirmedAt" IS NULL
  AND r."oldConfirmedAt" IS NOT NULL;

SELECT count(*) AS "انعاد تثبيتها" FROM "LegacyPendingReset" r
JOIN "Booking" b ON b.id = r."bookingId"
WHERE b."confirmedAt" = r."oldConfirmedAt" AND r."oldConfirmedAt" IS NOT NULL;

COMMIT;
SQL

echo "==> خلص. شغّل check-legacy-pending.sh مرة ثانية للتأكد."
