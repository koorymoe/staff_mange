package database

// ═══ صوت الموظفين — الفضفضة الأسبوعية + المشاكل الوظيفية (قرار (ع) 10-05) ═══
//
//   - PeerCheckin: كل أسبوع ماتركس يسأل الموظف شلونه وشي مضايقه.
//   - PeerNote: تقييمه لزميل (نجوم + كلمة). إذا التقييم سيء، ماتركس يتعمّق:
//     شنو صار؟ ليش؟ شلون تريد ينحل؟ ويقترح حل، والتقرير يروح للمراقب والمدير.
//     ⚠️ الزميل المقصود **ما يشوف أي شي** — ماكو مسار يرجّع إله.
//   - WorkplaceIssue: مشكلة وظيفية بين موظفين؛ ماتركس يسجّل ويقترح، والمراقب
//     والمدير يقررون، وبعدها ماتركس يتابع ويا الطرفين.
func peerVoiceMigrations() []Migration {
	return []Migration{{
		Version: "0320_peer_voice",
		SQL: `
			CREATE TABLE IF NOT EXISTS "PeerCheckin" (
				id           TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"employeeId" TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				week         DATE NOT NULL,
				mood         INTEGER CHECK (mood BETWEEN 1 AND 5),
				worry        TEXT,
				suggestion   TEXT,
				nothing      BOOLEAN NOT NULL DEFAULT false,
				urgent       BOOLEAN NOT NULL DEFAULT false,
				"createdAt"  TIMESTAMPTZ NOT NULL DEFAULT now(),
				"updatedAt"  TIMESTAMPTZ NOT NULL DEFAULT now(),
				UNIQUE ("employeeId", week)
			);
			CREATE TABLE IF NOT EXISTS "PeerNote" (
				id           TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"checkinId"  TEXT NOT NULL REFERENCES "PeerCheckin"(id) ON DELETE CASCADE,
				"authorId"   TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"targetId"   TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				score        INTEGER NOT NULL CHECK (score BETWEEN 1 AND 5),
				kind         TEXT NOT NULL DEFAULT 'NOTE' CHECK (kind IN ('GOOD', 'NOTE', 'PROBLEM')),
				text         TEXT,
				"needsFollowup" BOOLEAN NOT NULL DEFAULT false,
				"fWhat"      TEXT,
				"fWhy"       TEXT,
				"fWish"      TEXT,
				suggestion   TEXT,
				accepted     BOOLEAN,
				"acceptNote" TEXT,
				severity     TEXT NOT NULL DEFAULT 'NORMAL' CHECK (severity IN ('NORMAL', 'ATTENTION', 'URGENT')),
				"reportedAt" TIMESTAMPTZ,
				"createdAt"  TIMESTAMPTZ NOT NULL DEFAULT now(),
				CHECK ("authorId" <> "targetId")
			);
			CREATE INDEX IF NOT EXISTS "PeerNote_target_idx" ON "PeerNote" ("targetId", "createdAt" DESC);
			CREATE INDEX IF NOT EXISTS "PeerNote_author_idx" ON "PeerNote" ("authorId", "createdAt" DESC);

			CREATE TABLE IF NOT EXISTS "WorkplaceIssue" (
				id             TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				title          TEXT NOT NULL,
				"partyAId"     TEXT NOT NULL REFERENCES "Employee"(id) ON DELETE CASCADE,
				"partyBId"     TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"openedById"   TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				description    TEXT NOT NULL,
				severity       TEXT NOT NULL DEFAULT 'ATTENTION' CHECK (severity IN ('NORMAL', 'ATTENTION', 'URGENT')),
				status         TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'IN_PROGRESS', 'RESOLVED')),
				"aiSummary"    TEXT,
				decision       TEXT,
				"decidedById"  TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				"decidedAt"    TIMESTAMPTZ,
				"followUpAt"   TIMESTAMPTZ,
				"sourceNoteId" TEXT REFERENCES "PeerNote"(id) ON DELETE SET NULL,
				"createdAt"    TIMESTAMPTZ NOT NULL DEFAULT now(),
				"updatedAt"    TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE TABLE IF NOT EXISTS "WorkplaceIssueEntry" (
				id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
				"issueId"   TEXT NOT NULL REFERENCES "WorkplaceIssue"(id) ON DELETE CASCADE,
				"authorId"  TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
				kind        TEXT NOT NULL CHECK (kind IN ('STATEMENT', 'NOTE', 'DECISION', 'FOLLOWUP', 'MATRIX')),
				text        TEXT NOT NULL,
				"createdAt" TIMESTAMPTZ NOT NULL DEFAULT now()
			);
			CREATE INDEX IF NOT EXISTS "WorkplaceIssueEntry_issue_idx" ON "WorkplaceIssueEntry" ("issueId", "createdAt");`,
	}}
}
