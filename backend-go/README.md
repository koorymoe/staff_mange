# staffmange-api (Go)

باك إند جديد بلغة Go يحل تدريجياً محل الباك إند الحالي بـ TypeScript (`../backend`)،
اللي يبقى شغال بالتوازي حتى يكتمل الانتقال.

## الهيكل المعماري (Layered Architecture)

- `cmd/api/main.go` — نقطة تشغيل السيرفر، تحميل الإعدادات، ربط كل الطبقات
- `internal/config` — قراءة متغيرات البيئة
- `internal/database` — الاتصال بـ PostgreSQL عبر sqlx
- `internal/model` — الموديلات و DTOs
- `internal/repository` — طبقة SQL المباشرة
- `internal/service` — منطق العمل والتحقق (business logic)
- `internal/handler` — HTTP handlers وربط الـ endpoints
- `internal/middleware` — JWT auth, CORS, logging, recovery
- `migrations` — توثيق بنية الجداول (الجداول موجودة أصلاً بقاعدة Supabase)

## الحالة الحالية

✅ تم إنجازه:
- الهيكل الأساسي وربط الطبقات
- الاتصال بقاعدة البيانات (نفس قاعدة Supabase الحالية)
- توثيق JWT حقيقي (كان مفقوداً بالكامل بالباك إند القديم) + bcrypt
- Middleware: CORS, Logging, Recovery, RequireAuth, RequireRole
- أول Vertical Slice كامل: تسجيل الدخول + CRUD الموظفين (محمي بصلاحية ADMIN)

⏳ الباقي (بالترتيب المقترح):
1. الصلاحيات (Permissions)
2. الخدمات والمهارات (Services/Skills)
3. العملاء والحجوزات (Customers/Bookings)
4. المصاريف، الحضور، التقييم (KPI)
5. GPS، المشاريع، المشتريات، عروض الأسعار
6. رفع الصور (Cloudflare R2)
7. Server-Sent Events (لو احتجناها للتنبيهات الفورية)

## التشغيل محلياً

```bash
cp .env.example .env   # وحدّث DATABASE_URL و JWT_SECRET
go run ./cmd/api
```

## قواعد التطوير (يجب الالتزام بها لأي إضافة)

- كل جدول/تعديل بقاعدة البيانات → migration SQL موثق
- كل endpoint جديد → Model + Repository + Service + Handler (بدون منطق عمل داخل الـ handler)
- Response format ثابت: نجاح `{"data": ...}` / خطأ `{"error": "...", "code": N}`
- أي route محمي → يمر عبر `middleware.RequireAuth` (و`RequireRole` عند الحاجة)
- ممنوع وضع أسرار (JWT_SECRET, R2 keys) بالكود — فقط عبر متغيرات البيئة
