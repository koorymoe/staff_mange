package database

// ═══ رجوع الحجوزات القديمة لـ«بانتظار التثبيت» ═══
//
// «رجّعهن كلهن لبانتظار التثبيت».
//
// البوابة الجديدة (تواصل وية الزبون ← ثبّت ورحّل) انبنت بعد ما
// اشتغل النظام بأشهر. فالحجوزات الي انثبّتت قبلها عندها `confirmedAt`
// مسجّل، ومحد يعرف إذا أحد حچى وية زبونها فعلاً — فتطلع بمحطة «تم
// التثبيت» وكأنها عبرت البوابة وهي ما مرّت بيها.
//
// ⚠️ النطاق مضبوط بدقة — **ما نلمس شغلاً شغّالاً**:
//   - الي عليه كادر مكلّف → ما ينلمس (الفني يمكن طالع بيه اليوم).
//   - الي بدا التنفيذ (وصل/باشر/قيد التنفيذ) → ما ينلمس.
//   - المنجز والملغى → ما ينلمسون (تاريخ، مو طابور).
//   - المحبوس عند إدارة المشاريع → ما ينلمس (محطته الخاصة).
//   - الي انطلب حذفه → ما ينلمس (ينتظر قرار المراقب).
//
// يعني الي يرجع هو **بالضبط** الي كان قاعد بـ«تم التثبيت — بحاجة
// لكادر»: مثبّت على الورق، وماكو عليه ولا حركة.
//
// ⚠️ والرجوع **قابل للتراجع**: نحفظ القيم القديمة بجدول قبل ما
// نغيّرها. تعديل بالجملة على مئات الصفوف بلا نسخة يعني لو طلع الحكم
// غلط، ماكو طريق رجوع. للتراجع:
//
//	UPDATE "Booking" b SET "confirmedAt" = r."oldConfirmedAt",
//	  "confirmationContactedAt" = r."oldContactedAt", status = r."oldStatus"::"BookingStatus"
//	FROM "LegacyPendingReset" r WHERE r."bookingId" = b.id;
//
// ⚠️ و`confirmedByName`/`confirmedByEmployeeId` **ما ننظّفهن**: منو
// ثبّت الحجز أول مرة واقعة صارت، ومسحها يمحي تاريخاً ما إله بديل.
// المحطة تنحسب من `confirmedAt` وحده.
func legacyPendingResetMigrations() []Migration {
	return []Migration{
		{
			Version: "0250_legacy_pending_reset",
			SQL: `
				CREATE TABLE IF NOT EXISTS "LegacyPendingReset" (
					"bookingId" TEXT PRIMARY KEY REFERENCES "Booking"(id) ON DELETE CASCADE,
					"oldConfirmedAt" TIMESTAMPTZ,
					"oldContactedAt" TIMESTAMPTZ,
					"oldStatus" TEXT NOT NULL,
					"resetAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
				);

				-- نحفظ القديم أول (ON CONFLICT DO NOTHING حتى لو انعادت
				-- الهجرة ما تدوس على النسخة الأصلية بقيم بعد التعديل)
				INSERT INTO "LegacyPendingReset" ("bookingId", "oldConfirmedAt", "oldContactedAt", "oldStatus")
				SELECT b.id, b."confirmedAt", b."confirmationContactedAt", b.status::text
				FROM "Booking" b
				WHERE b."archivedAt" IS NULL
				  AND b."confirmedAt" IS NOT NULL
				  AND b.status NOT IN ('COMPLETED', 'CANCELLED', 'PARTIAL', 'IN_PROGRESS')
				  AND b."startedAt" IS NULL AND b."arrivedAt" IS NULL
				  AND NOT EXISTS (SELECT 1 FROM "BookingAssignment" ba WHERE ba."bookingId" = b.id)
				  AND NOT (b."transferToProjects" AND b."projectExecutionAt" IS NULL)
				  AND NOT EXISTS (SELECT 1 FROM "BookingDeleteRequest" dr
				                  WHERE dr."bookingId" = b.id AND dr.status = 'PENDING')
				ON CONFLICT ("bookingId") DO NOTHING;

				-- وبعدين نرجّعهن للطابور
				UPDATE "Booking" b SET
					"confirmedAt" = NULL,
					"confirmationContactedAt" = NULL,
					status = 'PENDING',
					"updatedAt" = now()
				FROM "LegacyPendingReset" r
				WHERE r."bookingId" = b.id AND b."confirmedAt" IS NOT NULL;
			`,
		},
	}
}
