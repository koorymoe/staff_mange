package integrity

// فحوصات النظام وماتركس (11-04). قراءة بس.
// مفاتيح المالك = model.PermissionOwnerOnly. حلقة متابعة التدريب كل ٦ ساعات على
// OPEN و followUpAt <= now (training_need.go). مطابقة التقرير: NO_TRACE ينتظر حكم المراقب.

func init() {
	Checks = append(Checks,
		Check{Key: "perm_owner_key_inactive_holder", Domain: "system", Severity: "high",
			Title:  "موظف مو فعّال وبعده شايل مفتاح مالك",
			Why:    "مفاتيح المالك (التدقيق المالي، إضافة وإيقاف الموظفين، محادثات الواتساب) تنمنح بالإيد — إذا رجع الحساب يرجع ويّاه المفتاح بلا ما أحد ينتبه.",
			Fix:    "نشيل المفتاح من الحساب الموقوف.",
			Count:  `SELECT count(*)::bigint n FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" JOIN "Employee" e ON e.id = ep."employeeId" WHERE p.name IN ('finance_audit','employee_create','employee_suspend','whatsapp_chats') AND e.status::text <> 'ACTIVE'`,
			Sample: `SELECT e.id, e.name||' ('||e.status::text||') — '||p.name info FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" JOIN "Employee" e ON e.id = ep."employeeId" WHERE p.name IN ('finance_audit','employee_create','employee_suspend','whatsapp_chats') AND e.status::text <> 'ACTIVE' LIMIT 20`},
		Check{Key: "training_need_followup_stuck", Domain: "system", Severity: "medium",
			Title:  "احتياج تدريب فات موعد متابعته وما انتابع",
			Why:    "ماتركس يتابع كل ٦ ساعات — إذا فات يومين معناها الحلقة واقفة، والموظف ما ينقاس إذا تحسّن.",
			Fix:    "نشوف سجل الخادم لحلقة التدريب ونشغّل المتابعة.",
			Count:  `SELECT count(*)::bigint n FROM "TrainingNeed" WHERE status = 'OPEN' AND "followUpAt" < now() - interval '2 days'`,
			Sample: `SELECT t.id, COALESCE(e.name,'?')||' — موعد المتابعة '||(t."followUpAt"::date)::text info FROM "TrainingNeed" t LEFT JOIN "Employee" e ON e.id = t."employeeId" WHERE t.status = 'OPEN' AND t."followUpAt" < now() - interval '2 days' ORDER BY t."followUpAt" LIMIT 20`},
		Check{Key: "training_need_inactive_employee", Domain: "system", Severity: "low",
			Title:  "احتياج تدريب مفتوح لموظف مو فعّال",
			Why:    "ينحسب بمشاكل الموظفين ويطلع بالتقارير وهو طالع من الشغل.",
			Fix:    "نسد الاحتياج.",
			Count:  `SELECT count(*)::bigint n FROM "TrainingNeed" t JOIN "Employee" e ON e.id = t."employeeId" WHERE t.status = 'OPEN' AND e.status::text <> 'ACTIVE'`,
			Sample: `SELECT t.id, e.name||' ('||e.status::text||')' info FROM "TrainingNeed" t JOIN "Employee" e ON e.id = t."employeeId" WHERE t.status = 'OPEN' AND e.status::text <> 'ACTIVE' LIMIT 20`},
		Check{Key: "report_match_unreviewed_backlog", Domain: "system", Severity: "medium",
			Title:  "تقارير «ماكو أثر» ما راجعها المراقب من أسبوعين",
			Why:    "هالحالة ما تنحسب على الموظف لحد ما المراقب يحكم — فتأخير المراجعة يعني الكذب ما ينحاسب والصادق ما ينصف.",
			Fix:    "المراقب يراجعها من شاشة تقارير الموظفين.",
			Count:  `SELECT count(*)::bigint n FROM "DailyReportMatch" WHERE status = 'NO_TRACE' AND "reviewVerdict" IS NULL AND day < current_date - 14`,
			Sample: `SELECT m.id, COALESCE(e.name,'?')||' — '||m.day::text info FROM "DailyReportMatch" m LEFT JOIN "Employee" e ON e.id = m."employeeId" WHERE m.status = 'NO_TRACE' AND m."reviewVerdict" IS NULL AND m.day < current_date - 14 ORDER BY m.day LIMIT 20`},
		Check{Key: "inline_files_left", Domain: "system", Severity: "low",
			Title:  "ملفات بعدها محفوظة داخل القاعدة (base64)",
			Why:    "قرار ٠٠٠٥: الملفات بمخزن الملفات والقاعدة بيها الرابط بس. السيرفر ينقل القديم لحاله بكل تشغيل (٥٠٠ لكل عمود)، فإذا الرقم ما ينزل معناها النقل يفشل — أو شاشة بعدها تحفظ base64.",
			Fix:    "نشوف سجل [storage] بالخادم. والمصفوفات (المعارض والفحوصات) بأداة cmd/migratefiles -apply.",
			Count:  `SELECT count(*)::bigint n FROM (SELECT 'Product.imageBase64' col, id::text id FROM "Product" WHERE "imageBase64" LIKE 'data:%' UNION ALL SELECT 'ProductProcurement.receiptImage' col, id::text id FROM "ProductProcurement" WHERE "receiptImage" LIKE 'data:%' UNION ALL SELECT 'RevolvingFundTxn.receiptImage' col, id::text id FROM "RevolvingFundTxn" WHERE "receiptImage" LIKE 'data:%' UNION ALL SELECT 'VehicleLog.receiptPhotoBase64' col, id::text id FROM "VehicleLog" WHERE "receiptPhotoBase64" LIKE 'data:%' UNION ALL SELECT 'VehicleDocument.fileUrl' col, id::text id FROM "VehicleDocument" WHERE "fileUrl" LIKE 'data:%' UNION ALL SELECT 'Project.contractPdfBase64' col, id::text id FROM "Project" WHERE "contractPdfBase64" LIKE 'data:%' UNION ALL SELECT 'Project.signedContractPdfBase64' col, id::text id FROM "Project" WHERE "signedContractPdfBase64" LIKE 'data:%' UNION ALL SELECT 'GpsDeviceRequest.invoicePhotoUrl' col, id::text id FROM "GpsDeviceRequest" WHERE "invoicePhotoUrl" LIKE 'data:%' UNION ALL SELECT 'Exhibition.businessCardPhotos' col, id::text id FROM "Exhibition" WHERE EXISTS (SELECT 1 FROM unnest("businessCardPhotos") x WHERE x LIKE 'data:%') UNION ALL SELECT 'ProjectChecklist.photoUrls' col, id::text id FROM "ProjectChecklist" WHERE EXISTS (SELECT 1 FROM unnest("photoUrls") x WHERE x LIKE 'data:%')) f`,
			Sample: `SELECT f.id, f.col info FROM (SELECT 'Product.imageBase64' col, id::text id FROM "Product" WHERE "imageBase64" LIKE 'data:%' UNION ALL SELECT 'ProductProcurement.receiptImage' col, id::text id FROM "ProductProcurement" WHERE "receiptImage" LIKE 'data:%' UNION ALL SELECT 'RevolvingFundTxn.receiptImage' col, id::text id FROM "RevolvingFundTxn" WHERE "receiptImage" LIKE 'data:%' UNION ALL SELECT 'VehicleLog.receiptPhotoBase64' col, id::text id FROM "VehicleLog" WHERE "receiptPhotoBase64" LIKE 'data:%' UNION ALL SELECT 'VehicleDocument.fileUrl' col, id::text id FROM "VehicleDocument" WHERE "fileUrl" LIKE 'data:%' UNION ALL SELECT 'Project.contractPdfBase64' col, id::text id FROM "Project" WHERE "contractPdfBase64" LIKE 'data:%' UNION ALL SELECT 'Project.signedContractPdfBase64' col, id::text id FROM "Project" WHERE "signedContractPdfBase64" LIKE 'data:%' UNION ALL SELECT 'GpsDeviceRequest.invoicePhotoUrl' col, id::text id FROM "GpsDeviceRequest" WHERE "invoicePhotoUrl" LIKE 'data:%' UNION ALL SELECT 'Exhibition.businessCardPhotos' col, id::text id FROM "Exhibition" WHERE EXISTS (SELECT 1 FROM unnest("businessCardPhotos") x WHERE x LIKE 'data:%') UNION ALL SELECT 'ProjectChecklist.photoUrls' col, id::text id FROM "ProjectChecklist" WHERE EXISTS (SELECT 1 FROM unnest("photoUrls") x WHERE x LIKE 'data:%')) f LIMIT 20`},
	)
}
