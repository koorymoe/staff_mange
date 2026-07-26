# مرجع الصلاحيات (Permissions Reference)

## ليش هذا الملف موجود

كتالوج الصلاحيات معرّف بمصدر واحد بالباك اند: `backend-go/internal/model/permission.go`
(`DefaultPermissions` و`RoleDefaultPermissions`). لكن عرضها بالواجهة مقسّمة لمجموعات (categories)
بشكل يدوي منفصل بـ `frontend/src/pages/PermissionsPage.tsx` (متغيّر `permissionGroups`). الملفين
منفصلين عن بعض ويقدروا يطلعوا من التزامن (drift) — وهذا صار فعلاً هالأسبوع: صلاحية جديدة انضافت
بالباك اند وما ظهرت بمجموعتها الصحيحة بصفحة الصلاحيات لغاية ما انضافت يدوياً بـ`permissionGroups`
(الواجهة فيها شبكة أمان تحطها بمجموعة "أخرى" تلقائياً بدل ما تختفي بصمت، لكن هذا مو بديل عن
تصنيفها صح بمجموعتها المنطقية).

**لما تضيف صلاحية جديدة، لازم تحدّث الثلاثة مع بعض:**
1. `backend-go/internal/model/permission.go` — أضفها بـ`DefaultPermissions` (وبـ
   `RoleDefaultPermissions` لو تبيها افتراضية لدور معيّن).
2. `frontend/src/pages/PermissionsPage.tsx` — أضفها بالمجموعة المناسبة جوا `permissionGroups`.
3. هذا الملف — أضفها بالجدول الأول تحت، وحدّث جدول الأدوار لو أضفتها كافتراضية لدور.

لو تشك فيه فرق موجود حالياً، قارن يدوياً بين قائمة `DefaultPermissions` وقائمة كل الأسماء المذكورة
جوا `permissionGroups` (بما فيها مجموعة "أخرى" التلقائية بالواجهة اللي تكشف أي صلاحية ناقصة تصنيف).

## جدول الصلاحيات

| الاسم (name) | التسمية بالعربي (label) | تُستخدم لحماية | ملاحظة |
|---|---|---|---|
| `staff_management` | إدارة الكوادر | مسارات إدارة الموظفين (middleware `requireStaffManagement`، صلاحية `staff_management`) | تحقق مباشر بـ `main.go` |
| `edit_employee_profile` | تعديل ملف الموظف (الراتب/الدوام/الإجازات) | تعديل بيانات حساسة بملف موظف (راتب/دوام/إجازات) | لا يوجد middleware مسار مستقل مخصص لها بـ`main.go` — تحقق ضمن منطق الميزة نفسها |
| `kpi_management` | تقييم الأداء (KPI) | مسارات تقييم الأداء (middleware `requireKpi`) | تحقق مباشر بـ `main.go` |
| `kpi_criteria_management` | إدارة نقاط الكي بي اي (إضافة/حذف) | إدارة معايير/نقاط تقييم الأداء (middleware `requireKpiCriteria`) | تحقق مباشر بـ `main.go` |
| `inventory` | جرد الأدوات | صفحة/عمليات جرد الأدوات | لا يوجد middleware مسار عام مخصص — تحقق ضمن منطق الميزة |
| `complaints` | الشكاوى | صفحة/عمليات الشكاوى | لا يوجد middleware مسار عام مخصص — تحقق ضمن منطق الميزة |
| `sales_booking` | إنشاء حجز جديد | إنشاء حجز جديد (مبيعات) | لا يوجد middleware مسار عام مخصص — تحقق ضمن منطق الميزة |
| `manage_customers` | إدارة العملاء | صفحة/عمليات إدارة العملاء | لا يوجد middleware مسار عام مخصص — تحقق ضمن منطق الميزة |
| `view_bookings` | عرض الحجوزات | عرض قوائم الحجوزات | لا يوجد middleware مسار عام مخصص — تحقق ضمن منطق الميزة |
| `coordinator` | تنسيق الحجوزات | تنسيق/توزيع الحجوزات (صفحة `Coordinator.tsx`) | لا يوجد middleware مسار عام مخصص — تحقق ضمن منطق الميزة |
| `manage_services` | الخدمات | إدارة الخدمات المقدَّمة | لا يوجد middleware مسار عام مخصص — تحقق ضمن منطق الميزة |
| `mission_tracking` | تتبع المهام | تتبع المهام الميدانية | لا يوجد middleware مسار عام مخصص — تحقق ضمن منطق الميزة |
| `gps_system` | نظام GPS | كل مسارات نظام GPS (middleware `requireGpsSystem`) | تحقق مباشر بـ `main.go`؛ تُمنح تلقائياً لكل موظف بدور MONITOR (شوف `grantGpsSystemToMonitors` بـ`seed_accounts.go`) |
| `project_management` | إدارة المشاريع | مسارات إدارة المشاريع (middleware `requireProjectMgmtPerm`) | تحقق مباشر بـ `main.go` |
| `quotation_system` | نظام عروض الأسعار | مسارات عروض الأسعار (middleware `requireQuotationSystem`) | تحقق مباشر بـ `main.go` |
| `finance` | المالية | التحقق من الحجوزات مالياً وعمليات مالية أخرى (middleware `requireVerifyBooking`) | تحقق مباشر بـ `main.go` |
| `expenses` | المصاريف | صفحة/عمليات المصاريف | لا يوجد middleware مسار عام مخصص — تحقق ضمن منطق الميزة |
| `procurement` | المشتريات | مسارات المشتريات العامة (middleware `requireProcurement`) | تحقق مباشر بـ `main.go` |
| `procurement_personal` | طلب احتياجات شخصية | طلب مشتريات لاحتياجات شخصية للموظف | لا يوجد middleware مسار عام مخصص — تحقق ضمن منطق الميزة |
| `procurement_customer` | طلب منتج للزبون | طلب مشتريات نيابة عن زبون | لا يوجد middleware مسار عام مخصص — تحقق ضمن منطق الميزة |
| `monitoring` | مراقبة (متابعة المهام والحجوزات والموظفين) | صلاحيات المراقبة العامة (لوحة المراقب) | لا يوجد middleware مسار عام مخصص — تحقق ضمن منطق الميزة |
| `auditing` | تدقيق (التحقق من جودة العمل والتقارير والحسابات) | صلاحيات التدقيق | لا يوجد middleware مسار عام مخصص — تحقق ضمن منطق الميزة |
| `content_technician` | صلاحية التقني (إدارة المحتوى التدريبي والخدمات والموردين والمواد) | مسارات إدارة المحتوى التدريبي/الخدمات/الموردين/المواد (middleware `requireContentTech`) | تحقق مباشر بـ `main.go` |
| `vehicle_management` | إدارة المركبات | كل مسارات إدارة الأسطول (middleware `requireVehicleMgmt`) | تحقق مباشر بـ `main.go` |
| `quality_control` | الجودة (متابعة مشاكل التنفيذ والرقابة) | مسارات الجودة (middleware `requireQuality`) | تحقق مباشر بـ `main.go` |

> ملاحظة على العمود الأخير: الصلاحيات المذكور فيها middleware صريح بـ`main.go` تُفرض على مستوى
> المسار (route) مباشرة عن طريق `RequirePermission`. البقية غالباً تُتحقق منها الواجهة/منطق الخدمة
> نفسه (أو تُستخدم لإظهار/إخفاء عناصر بالواجهة) بدل بوابة `RequirePermission` مستقلة على مسار كامل —
> هذا لا يعني إنها غير فعّالة، بس آلية التحقق منها مختلفة عن الصلاحيات المذكور أمامها middleware.

## جدول الأدوار والصلاحيات الافتراضية

من `RoleDefaultPermissions` بـ `permission.go` — هذي الصلاحيات تُمنح تلقائياً لأي موظف حسب دوره
الوظيفي وقت إنشائه (أو تُمنح رجعياً للموجودين لو تغيّرت الخارطة، شوف `grantRolePermission` بـ
`seed_accounts.go`).

| الدور (role) | الصلاحيات الافتراضية |
|---|---|
| `SALES` | `sales_booking`, `complaints` |
| `HR_COORDINATOR` | `staff_management`, `edit_employee_profile`, `coordinator`, `manage_customers`, `view_bookings`, `manage_services`, `inventory`, `complaints`, `mission_tracking`, `sales_booking` |
| `TECHNICIAN` | `expenses` |
| `PROJECT_MANAGER` | `project_management`, `expenses`, `mission_tracking` |
| `MONITOR` | `staff_management`, `edit_employee_profile`, `kpi_management`, `view_bookings`, `manage_customers`, `manage_services`, `mission_tracking`, `inventory`, `complaints`, `finance`, `monitoring`, `auditing`, `quality_control`, `gps_system` |
| `FINANCE` | `finance`, `view_bookings` |
| `GPS_ADMIN` | `gps_system` |
| `QUALITY_ENGINEER` | `auditing`, `complaints`, `quality_control`, `sales_booking`, `kpi_management` |
| `ENGINEER` | `expenses`, `quotation_system`, `project_management` |
| `PROCUREMENT_ADMIN` | `procurement`, `inventory` |
| `ADMIN` | *(فاضية — شوف الملاحظة تحت)* |
| `DESIGNER` | *(فاضية — شوف الملاحظة تحت)* |
| `SERVICE_MANAGER` | *(فاضية — شوف الملاحظة تحت)* |

> **مهم**: `ADMIN`, `DESIGNER`, `SERVICE_MANAGER` عندهم قائمة صلاحيات افتراضية فاضية بالكود. هذا
> منطقي بالنسبة لـ`ADMIN` لأنه أصلاً يتخطى كل فحص `RequirePermission` تلقائياً (شوف
> `internal/middleware/auth.go` — `ADMIN`/`OWNER` يمرّون دائماً بدون حاجة لصلاحية محددة). أما
> `DESIGNER` و`SERVICE_MANAGER` فما عندهم أي تخطي تلقائي كهذا — صلاحياتهم الفعلية لازم تُمنح يدوياً
> لكل موظف على حدة من صفحة الصلاحيات (`PermissionsPage.tsx`)، ما فيه أي منح تلقائي وقت الإنشاء.
> حساب `OWNER` غير مذكور بـ`RoleDefaultPermissions` أصلاً لأنه أعلى من كل هذا النظام بالكامل —
> يتخطى `RequireRole` و`RequirePermission` معاً دائماً (شوف `ARCHITECTURE.md`).
