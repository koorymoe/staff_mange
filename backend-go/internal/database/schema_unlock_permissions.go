package database

// ══════════════════════════════════════════════════════════════════
// صلاحيات تفتح الشاشات الي ما كانت تنفتح بأي منح
// ══════════════════════════════════════════════════════════════════
//
// شكوى صاحب العمل حرفياً: «صلاحية من أنطيها لأحد يلا تظهر إله».
// والسبب إن ٣١ شاشة بالنظام كانت مشروطة **بالدور بس** وماكو إلها
// صلاحية أصلاً — يعني المدير ينطي شنو ما ينطي، الشاشة تضل مخفية
// لأنه ماكو مفتاح يفتحها.
//
// هاي الصلاحيات هي المفاتيح الناقصة. كل وحدة تفتح شاشة كانت مقفولة
// على دور معيّن، فيقدر المالك يفوّض الشغل لأي موظف بلا ما يغيّر دوره.
//
// ⚠️ ما تنمنح لأحد تلقائياً. الأدوار تشتغل مثل ما هي بالضبط — هاي
// إضافة فوقها مو بديل عنها، فما ينكسر ولا وصول موجود.
func unlockPermissionsMigration() []Migration {
	return []Migration{
		{
			Version: "0239_unlock_permissions",
			SQL: `
				INSERT INTO "Permission" (id, name, label) VALUES
					(gen_random_uuid()::text, 'my_projects',        'المشاريع الموجّهة لي'),
					(gen_random_uuid()::text, 'work_reports',       'تقارير العمل'),
					(gen_random_uuid()::text, 'leader_invoices_view', 'عرض فواتير الليدر'),
					(gen_random_uuid()::text, 'expenses_manage',    'إدارة المصاريف'),
					(gen_random_uuid()::text, 'audit_issues',       'بلاغات أخطاء التدقيق'),
					(gen_random_uuid()::text, 'gps_install_costs',  'حساب تكاليف الشد'),
					(gen_random_uuid()::text, 'device_maintenance', 'صيانة الأجهزة'),
					(gen_random_uuid()::text, 'performance_review', 'تقييم الأداء'),
					(gen_random_uuid()::text, 'staff_requests',     'طلبات الكادر'),
					(gen_random_uuid()::text, 'employee_stats',     'إحصائيات الموظفين'),
					(gen_random_uuid()::text, 'service_managers',   'مسؤولو الخدمات'),
					(gen_random_uuid()::text, 'network_prices',     'تعديل أسعار الشبكات'),
					(gen_random_uuid()::text, 'create_booking',     'إنشاء حجز جديد'),
					(gen_random_uuid()::text, 'design_forms',       'فورمة التصميم'),
					(gen_random_uuid()::text, 'gps_requests',       'طلبات الجي بي اس')
				ON CONFLICT (name) DO NOTHING;
			`,
		},
	}
}
