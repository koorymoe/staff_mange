package model

// (دازاي 11-28) «لازم كل شخص صلاحياته حسب المنصب مالته… نائب مدير الادارة المفروض يشوف التقييمات
// والاحصائيات ويشوف التفاصيل مال الموظفين ويدقق شغلهم… وهذا ينطبق ع كل المناصب والادارات»:
//   - النائب (وحدة أو إدارة): صلاحيات المدير الي ما عنده **مشاهدة بس** — يشوف كلشي، والموافقة والتعديل للمدير.
//   - مدير/نائب «الإدارة» (admin): فوگها مشاهدة كل شاشات الرقابة والتدقيق (OversightBundle).
// والمناصب المتعددة تتجمّع (مدير سلاسل التوريد + نائب الإدارة = الاثنين).

// OrgViewSQL صلاحيات «مشاهدة» من المنصب لموظف ($1). صلاحيات المالك مستثناة بالكود.
const OrgViewSQL = `
	SELECT DISTINCT b.permission FROM "OrgAssignment" a
	JOIN "OrgBundle" b ON b.scope = a."unitId"
	WHERE a."employeeId" = $1 AND a."endedAt" IS NULL AND a."unitId" IS NOT NULL AND a.position = 'DEPUTY' AND b."headOnly"
	UNION
	SELECT DISTINCT b.permission FROM "OrgAssignment" a
	JOIN "OrgBundle" b ON b.scope = a."divisionId"
	WHERE a."employeeId" = $1 AND a."endedAt" IS NULL AND a."unitId" IS NULL AND a.position = 'DEPUTY' AND b."headOnly"
	UNION
	SELECT DISTINCT b.permission FROM "OrgAssignment" a
	JOIN "OrgUnit" u ON u."divisionId" = a."divisionId"
	JOIN "OrgBundle" b ON b.scope = u.id
	WHERE a."employeeId" = $1 AND a."endedAt" IS NULL AND a."unitId" IS NULL AND a.position = 'DEPUTY' AND b."headOnly"`

// OrgOversightSQL هل الموظف ($1) مدير أو نائب «الإدارة»؟
const OrgOversightSQL = `SELECT EXISTS (SELECT 1 FROM "OrgAssignment" a WHERE a."employeeId" = $1 AND a."endedAt" IS NULL
	AND a."unitId" IS NULL AND a."divisionId" = 'admin' AND a.position IN ('DIRECTOR', 'DEPUTY'))`
