-- فحص صحة البيانات — قراية بس (READ ONLY + ROLLBACK)، ما يغيّر شي.
-- التشغيل على السيرفر:
--   docker compose exec -T db psql -U staffmange -d staffmange < docs/data-health.sql
BEGIN READ ONLY;
SELECT 'حجوزات خلصت بلا فاتورة (60 يوم)' AS "الفحص", count(*)::text AS "العدد" FROM "Booking" b
 WHERE b.status='COMPLETED' AND b."archivedAt" IS NULL AND b."settledLegacyAt" IS NULL
   AND b."completedAt" >= (now() AT TIME ZONE 'UTC') - interval '60 days'
   AND NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId"=b.id)
UNION ALL SELECT 'فواتير تنتظر التدقيق > يومين', count(*)::text FROM "LeaderInvoice"
 WHERE status='SUBMITTED' AND COALESCE(btrim("auditVerdict"),'')='' AND "createdAt" < (now() AT TIME ZONE 'UTC') - interval '2 days'
UNION ALL SELECT 'فواتير مدققة تنتظر الاعتماد > يومين', count(*)::text FROM "LeaderInvoice"
 WHERE status='SUBMITTED' AND COALESCE(btrim("auditVerdict"),'')<>'' AND "auditedAt" < now() - interval '2 days'
UNION ALL SELECT 'فواتير مسحوبة الاعتماد (تختفي من المجاميع)', count(*)::text FROM "LeaderInvoice" WHERE "revokedAt" IS NOT NULL
UNION ALL SELECT 'حجوزات بيها أكثر من فاتورة', count(*)::text FROM (SELECT "bookingId" FROM "LeaderInvoice" WHERE "revokedAt" IS NULL AND "bookingId" IS NOT NULL GROUP BY 1 HAVING count(*)>1) t
UNION ALL SELECT 'دفعات مشاريع ما تأكدت (عدد)', count(*)::text FROM "ProjectPayment" WHERE "cancelledAt" IS NULL AND "verifiedAt" IS NULL
UNION ALL SELECT 'دفعات مشاريع ما تأكدت (مبلغ)', COALESCE(sum(amount),0)::text FROM "ProjectPayment" WHERE "cancelledAt" IS NULL AND "verifiedAt" IS NULL
UNION ALL SELECT 'مشاريع بلا سعر (بلا المرفوض)', count(*)::text FROM "Project" p
 WHERE COALESCE(p.stage,'') NOT LIKE '%مرفوض%' AND NULLIF(regexp_replace(COALESCE(p.price,''),'\D','','g'),'') IS NULL
   AND NOT EXISTS (SELECT 1 FROM "ProjectContractValue" cv WHERE cv."projectId"=p.id)
UNION ALL SELECT 'مشاريع بلا مسؤول (بلا المرفوض)', count(*)::text FROM "Project" p
 WHERE COALESCE(p.stage,'') NOT LIKE '%مرفوض%' AND p."responsibleEmployeeId" IS NULL
UNION ALL SELECT 'حجوزات عالقة بعلم الانتظار', count(*)::text FROM "Booking" WHERE "waitingSince" IS NOT NULL AND status NOT IN ('WAITING','CANCELLED')
UNION ALL SELECT 'حجوزات بيها أكثر من «تم» (إنجاز مكرر)', count(*)::text FROM (SELECT "bookingId" FROM "BookingVisit" WHERE outcome='DONE' GROUP BY 1 HAVING count(*)>1) t
UNION ALL SELECT 'زبائن مكررين بنفس الرقم بصيغة ثانية', count(*)::text FROM (
  SELECT right(regexp_replace(phone,'\D','','g'),10) k FROM "Customer" WHERE phone IS NOT NULL AND length(regexp_replace(phone,'\D','','g'))>=10 GROUP BY 1 HAVING count(*)>1) t
UNION ALL SELECT 'حجوزات مفتوحة مكلّف بيها موظف مو فعّال', count(*)::text FROM "BookingAssignment" a JOIN "Employee" e ON e.id=a."employeeId" JOIN "Booking" b ON b.id=a."bookingId"
 WHERE e.status<>'ACTIVE' AND b.status NOT IN ('COMPLETED','CANCELLED') AND b."archivedAt" IS NULL
UNION ALL SELECT 'فواتير على حجوزات ملغية', count(*)::text FROM "LeaderInvoice" li JOIN "Booking" b ON b.id=li."bookingId" WHERE li."revokedAt" IS NULL AND b.status='CANCELLED';
ROLLBACK;
