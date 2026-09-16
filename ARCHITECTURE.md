# نظرة معمارية على النظام (Architecture Overview)

هذا الملف موجّه لأي مطوّر (أو وكيل AI) جديد على المشروع — يشرح البنية العامة خلال ٥ دقائق قراءة
حتى تعرف وين تدوّر على أي شي. المعلومات هنا مبنية على قراءة الكود الفعلي، لا تخمين.

## التقنيات المستخدمة (Tech Stack)

- **Backend**: Go + [sqlx](https://github.com/jmoiron/sqlx) (بدون ORM كامل، استعلامات SQL مباشرة
  مع mapping للـ structs) + PostgreSQL 16. المصدر: `backend-go/`.
- **Frontend**: React + TypeScript + Vite. المصدر: `frontend/src/`.
- **النشر**: Docker Compose، مع Caddy كـ reverse proxy لتوفير HTTPS تلقائي (Let's Encrypt) —
  شوف `docker-compose.yml` و`Caddyfile` و`DEPLOY.md`.

## طبقات الباك اند: Handler → Service → Repository → Postgres

كل ميزة بالباك اند مقسّمة لثلاث طبقات بنفس النمط، كل واحدة بمجلدها الخاص تحت `backend-go/internal/`:

- **Handler** (`internal/handler/*.go`): يقرأ الـ HTTP request (body/params)، يستدعي دالة بالـ Service
  المناسب، ويكتب الـ JSON response. ما يحتوي منطق أعمال ولا استعلامات SQL.
- **Service** (`internal/service/*.go`): منطق الأعمال (business logic) — تحقق من صحة البيانات،
  قرارات، تنسيق بين أكثر من repository لو احتاج الأمر. يستدعي الـ Repository المناسب.
- **Repository** (`internal/repository/*.go`): طبقة الوصول لقاعدة البيانات فقط — استعلامات SQL عن
  طريق `sqlx` مباشرة على `*sqlx.DB`.
- التوصيل بين الطبقات ونقاط الـ API (المسارات وسلاسل الـ middleware) كلها مركّبة يدوياً بملف واحد:
  `backend-go/cmd/api/main.go`.

### مثال حقيقي: ميزة المركبات (Vehicles)

- `internal/model/vehicle.go`: تعريف `Vehicle` والـ request/response structs.
- `internal/repository/vehicle_repository.go`: `List()`, `Create()`, `Get()`, `Update()`,
  `ListDocuments()`... — كل دالة تنفذ SQL مباشرة على `*sqlx.DB`.
- `internal/service/vehicle_service.go`: `VehicleService` يغلّف `VehicleRepository`؛ أغلب الدوال
  (`List`, `Get`) تمرر مباشرة للـ repository، بينما `Create`/`Update` فيها منطق إضافي بسيط.
- `internal/handler/vehicle_handler.go`: `VehicleHandler` (`List`, `Create`, `Update`,
  `ListDocuments`...) يستقبل الـ HTTP request ويستدعي `VehicleService`.
- التوصيل والمسارات بـ `cmd/api/main.go`: `vehicleRepo := repository.NewVehicleRepository(db)` ثم
  `vehicleService := service.NewVehicleService(vehicleRepo)` ثم
  `vehicleHandler := handler.NewVehicleHandler(vehicleService)`، ثم تسجيل المسارات مثل
  `mux.Handle("GET /api/vehicles", middleware.Chain(http.HandlerFunc(vehicleHandler.List), requireAuth, requireVehicleMgmt))`
  — لاحظ سلسلة الـ middleware: أول `requireAuth` (تحقق الهوية)، بعدين `requireVehicleMgmt`
  (صلاحية `vehicle_management`، شوف قسم الصلاحيات تحت).

نفس النمط بالضبط ينطبق على بقية الميزات (حجوزات، عملاء، مشتريات، GPS...) — لو تدوّر على ميزة معينة،
دوّر على اسمها بـ `internal/model/`, `internal/repository/`, `internal/service/`, `internal/handler/`
وبعدين على مساراتها بـ `cmd/api/main.go`.

## المصادقة والصلاحيات (Auth & Permissions)

- **JWT**: تسجيل الدخول وإصدار/فك تشفير التوكن بـ `internal/service/auth_service.go`.
- **`RequireAuth`** (`internal/middleware/auth.go`): يتحقق من وجود توكن JWT صالح، لكن **لا يثق
  بالدور (role) المخزّن جوا التوكن نفسه** — بعد ما يتحقق من التوكن، يستعلم من قاعدة البيانات مباشرة
  عن الحالة (`status`) والدور (`role`) الحاليين للموظف (`employees.StatusAndRoleByID`) ويستخدمهم
  بكل فحوصات الصلاحيات اللاحقة. هذا تعديل متعمد: قبله، تنزيل موظف من ADMIN لدور عادي (أو إيقاف
  حسابه) ما كان يبطل صلاحياته الفعلية إلا بعد انتهاء صلاحية التوكن (لغاية ١٢ ساعة) أو تسجيل خروج
  ودخول يدوي — يعني موظف مفصول أو متراجع صلاحياته يقدر يستمر يستخدم النظام بصلاحياته القديمة لفترة.
  الحل: كل طلب يعيد التحقق من قاعدة البيانات فوراً.
- **`RequireRole`**: يقيّد المسار على أدوار محددة (`role` من الـ context اللي حطه `RequireAuth`).
  حساب `OWNER` يتخطى أي قيد أدوار دائماً.
- **`RequirePermission`**: تحقق دقيق (fine-grained) — يسمح دائماً لـ `ADMIN`/`OWNER`، وإلا يتحقق هل
  عند الموظف الصلاحية المطلوبة بجدول `EmployeePermission` (مرتبط بجدول `Permission` عن طريق
  `PermissionRepository.ListForEmployee`). هذا النظام يسمح تخصيص صلاحيات لموظف واحد بمعزل عن دوره
  الوظيفي (شوف `PERMISSIONS.md` للتفاصيل الكاملة عن كل صلاحية والأدوار الافتراضية).
- كلا النوعين (`RequireRole` و`RequirePermission`) موجودين بـ `internal/middleware/auth.go`، وأي
  محاولة وصول مرفوضة تُسجَّل وتنبّه الإدارة (ADMIN) بعد عدد معيّن من المحاولات المتكررة (بدون إيقاف
  تلقائي للحساب — قرار بشري).

## ترحيلات قاعدة البيانات (Migrations)

- كل تعديل بنية قاعدة البيانات مسجّل كـ "migration" برقم إصدار (version) فريد بجدول تتبّع اسمه
  `SchemaMigration` (عمود `version` هو المفتاح الأساسي).
- المنطق بـ `internal/database/migrate.go` — الدالة `runVersionedMigrations`: لكل migration، تتحقق
  هل رقمه مسجّل مسبقاً بجدول `SchemaMigration`؛ لو مسجّل تتخطاه بالكامل بدون تنفيذ أي SQL. لو غير
  مسجّل، تنفّذه جوا معاملة (transaction) مستقلة خاصة فيه، وتسجّل رقمه بنفس المعاملة، ثم commit. لو
  فشل أي migration يصير rollback فوري ويتوقف كل الترحيل بذيك اللحظة (ما يكمل بعد أول فشل).
- ترتيب الـ migrations ثابت ومرقّم تلقائياً حسب ترتيبه بقائمة الكود — أول واحد `0001_initial_schema`
  (كل بنية القاعدة الأساسية دفعة وحدة)، وبعده `0002`, `0003`... بنفس ترتيب مصفوفة `migrations` بملف
  `internal/database/schema_migrations.go` وأسمائها الوصفية بـ `migrationNames` بملف
  `internal/database/schema_versions.go`.
- **لإضافة migration جديد**: أضف عبارة SQL الجديدة بنهاية مصفوفة `migrations` (بـ
  `schema_migrations.go`) وأضف اسمها الوصفي بنهاية `migrationNames` (بـ `schema_versions.go`) —
  بنفس الترتيب المتقابل بين الملفين. **ممنوع منعاً باتاً تعديل أو حذف أو إعادة ترتيب أي عنصر موجود
  مسبقاً** بأي من المصفوفتين — هذا الترتيب هو رقم الإصدار الفعلي المسجّل بقواعد بيانات كل السيرفرات
  الشغّالة، وأي تغيير رجعي يكسر التزامن معهم.

## طوبولوجيا النشر (Deployment Topology)

`docker-compose.yml` يعرّف أربع خدمات:

- **`db`**: PostgreSQL 16 (alpine)، بيانات دائمة على volume اسمه `db_data`.
- **`backend`**: يبني من `backend-go/`، يتصل بـ`db` (`depends_on: service_healthy`)، يقرأ كل
  إعداداته من متغيرات بيئة (`DATABASE_URL`, `JWT_SECRET`, `CORS_ORIGIN`, `OWNER_USERNAME/PASSWORD`،
  إلخ — شوف `internal/config/config.go`). مجلد `tutorials/` (أدلة الاستخدام اللي يقرأها المساعد
  الذكي) موصول كـ bind-mount للتحديث بدون إعادة بناء الصورة.
- **`frontend`**: يبني من `frontend/`، ينتظر `backend` يصير `healthy` فعلياً (مو بس started) قبل
  ما يشتغل، ويكشف بورت `8081:80` داخلياً.
- **`caddy`**: reverse proxy أمام `frontend`، يكشف بورت `80`/`443` للعالم الخارجي، ويتكفل بشهادة
  HTTPS تلقائية عن طريق `DOMAIN` (شوف `DEPLOY.md` وقسم "الدومين" فيه).

**النسخ الاحتياطي (backups)**: `backup-db.sh` يسوي `pg_dump` يومي مضغوط (`.sql.gz`) بمجلد
`backups/` محلياً على نفس السيرفر، ويحذف تلقائياً أي نسخة أقدم من 14 يوم. `setup-backups.sh` يفعّل
هذا كـ cron job يومي (الساعة 3 فجراً). **قيود معروفة حالياً (فجوة موثّقة)**: النسخ محلية فقط على نفس
السيرفر (لا يوجد نسخ خارجي/offsite)، والاحتفاظ محدود بـ 14 يوم بس — لو انهار السيرفر بالكامل (قرص
تالف، حذف عرضي للـ volume) تنفقد كل النسخ الاحتياطية معه. يُنصح بإضافة رفع دوري لمكان خارجي (S3/R2
أو مشابه) مستقبلاً.

## خارطة "وين ألقى الأشياء" (Where Things Live)

| الميزة/المجال | Model | Repository | Service | Handler | صفحة الفرونت إند الرئيسية |
|---|---|---|---|---|---|
| الحجوزات (Bookings) | `internal/model/booking.go` | `internal/repository/booking_repository.go` | `internal/service/booking_service.go` | `internal/handler/booking_handler.go` | `frontend/src/pages/BookingsList.tsx`, `Coordinator.tsx`, `SalesBooking.tsx` |
| المركبات (Vehicle Fleet) | `internal/model/vehicle.go` | `internal/repository/vehicle_repository.go` | `internal/service/vehicle_service.go` | `internal/handler/vehicle_handler.go` | `frontend/src/pages/VehiclesPage.tsx`, `FleetDashboardPage.tsx` |
| مهام/حجوزات المركبات | — | `vehicle_mission_repository.go`, `vehicle_booking_repository.go` | `vehicle_mission_service.go`, `vehicle_booking_service.go` | `vehicle_mission_handler.go`, `vehicle_booking_handler.go` | `frontend/src/pages/VehicleMissionsPage.tsx` |
| نظام GPS | `internal/model/gps.go` | `internal/repository/gps_repository.go` | `internal/service/gps_service.go` | `internal/handler/gps_handler.go` | `frontend/src/pages/gps/` (`GpsDashboard.tsx`, `GpsCustomers.tsx`, `GpsDevices.tsx`, `GpsSims.tsx` وغيرها) |
| المشتريات (Procurement) | `internal/model/procurement.go` | `internal/repository/procurement_repository.go` | `internal/service/procurement_service.go` | `internal/handler/procurement_handler.go` | `frontend/src/pages/ProcurementPage.tsx` |
| الحضور (Attendance) | `internal/model/attendance.go` | `internal/repository/attendance_repository.go` | `internal/service/attendance_service.go` | `internal/handler/attendance_handler.go` | `frontend/src/pages/AttendancePage.tsx` |
| المساعد الذكي (AI Assistant) | `internal/model/assistant_conversation.go`, `assistant_knowledge.go` | `assistant_conversation_repository.go`, `assistant_knowledge_repository.go` | `internal/service/assistant_service.go` | `internal/handler/assistant_handler.go` | `frontend/src/pages/AssistantConversationsPage.tsx` |
| تقييم الأداء (KPI) | `internal/model/kpi.go`, `smart_kpi.go` | `kpi_repository.go`, `kpi_criterion_repository.go`, `smart_kpi_repository.go` | `kpi_service.go`, `kpi_criterion_service.go`, `smart_kpi_service.go` | `kpi_handler.go`, `kpi_criterion_handler.go`, `smart_kpi_handler.go` | `frontend/src/pages/KpiPage.tsx` |
| الصلاحيات (Permissions) | `internal/model/permission.go` | `internal/repository/permission_repository.go` | `internal/service/permission_service.go` | `internal/handler/permission_handler.go` | `frontend/src/pages/PermissionsPage.tsx` |
| الموظفين (Employees) | `internal/model/employee.go` | `internal/repository/employee_repository.go` | `internal/service/employee_service.go` | `internal/handler/employee_handler.go` | `frontend/src/pages/Employees.tsx` |
| العملاء (Customers) | `internal/model/customer.go` | `internal/repository/customer_repository.go` | `internal/service/customer_service.go` | `internal/handler/customer_handler.go` | `frontend/src/pages/Customers.tsx` |
| الجودة (Quality) | — | `quality_repository.go`, `quality_follow_up_repository.go` | `quality_service.go`, `quality_follow_up_service.go` | `quality_handler.go`, `quality_follow_up_handler.go` | `frontend/src/pages/QualityPage.tsx`, `QualityFollowUpsPage.tsx` |

نقطة تركيب كل شي (الاتصال بين الطبقات + تعريف المسارات + سلاسل الـ middleware): `backend-go/cmd/api/main.go`.
