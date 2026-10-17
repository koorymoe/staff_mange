package database

// ═══ دفعات المشاريع (قرار (ع) 10-06) ═══
// «من مشروع واحد محصل 500 مليون ليش ماموجوده بالاحصائيات» — النظام ما چان
// بيه مكان لفلوس المشروع أصلاً (السعر نص بس). هسه كل دفعة تنسجّل بمبلغها
// وتاريخها ووصلها؛ المحاسب ومدير المشاريع ومشرف المشروع يسجّلون، والمحاسب
// يأكد. وقيمة العقد رقم بجدول منفصل حتى "Project" (SELECT *) ما يتغيّر.
func projectPaymentMigrations() []Migration {
	return []Migration{{
		Version: "0323_project_payments",
		SQL: `
			CREATE TABLE IF NOT EXISTS "ProjectPayment" (
				id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"projectId"     TEXT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
				amount          NUMERIC(16, 0) NOT NULL CHECK (amount > 0),
				"paidAt"        DATE NOT NULL,
				method          TEXT NOT NULL DEFAULT 'CASH',   -- CASH | TRANSFER | CHEQUE
				"receiptNo"     TEXT,
				note            TEXT,
				"createdById"   TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"createdAt"     TIMESTAMPTZ NOT NULL DEFAULT now(),
				"verifiedById"  TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"verifiedAt"    TIMESTAMPTZ,
				"cancelledAt"   TIMESTAMPTZ,
				"cancelledById" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"cancelReason"  TEXT
			);
			CREATE INDEX IF NOT EXISTS "ProjectPayment_project_idx" ON "ProjectPayment"("projectId");
			CREATE INDEX IF NOT EXISTS "ProjectPayment_paid_idx" ON "ProjectPayment"("paidAt");
			CREATE TABLE IF NOT EXISTS "ProjectContractValue" (
				"projectId" TEXT PRIMARY KEY REFERENCES "Project"(id) ON DELETE CASCADE,
				amount      NUMERIC(16, 0) NOT NULL CHECK (amount >= 0),
				"setById"   TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"setAt"     TIMESTAMPTZ NOT NULL DEFAULT now()
			);
		`,
	}}
}
