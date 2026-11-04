# نظام شركة الأماني — staff_mange

نظام إدارة الشركة: الحجوزات والكوادر، والفلوس (فواتير الليدرية، والمصاريف، والدوار، ودفعات المشاريع)، والمشاريع، والمخازن والمركبات، والهيكل التنظيمي، وماتركس (التقييم والمتابعة بالذكاء). وويّاه تطبيق الزبائن.

| الجزء | المكان | التقنية |
|---|---|---|
| السيرفر | `backend-go/` | Go + sqlx + PostgreSQL |
| الواجهة | `frontend/` | React + TypeScript (Vite) |
| تطبيق الزبائن | `customer-app/` | React + TypeScript (Vite) |
| النشر | `docker-compose.yml`، `Caddyfile`، `deploy.sh` | Docker + Caddy |

## الفروع
- **`avg`**: فرع الشغل. كل تعديل ينرفع هنا أول.
- **`prod`**: النسخة الي يشتغل عليها الموظفين. ما ينرفع عليه إلا بطلب صريح من صاحب النظام.

## التشغيل محلياً
```bash
# السيرفر (يطبّق الترحيلات لحاله بأول تشغيل)
cd backend-go
DATABASE_URL=postgres://USER:PASS@localhost:5432/DB?sslmode=disable \
JWT_SECRET='سر-طوله-٣٢-حرف-أو-أكثر' OWNER_USERNAME=owner OWNER_PASSWORD='...' \
PORT=4000 CORS_ORIGIN=http://127.0.0.1:5199 go run ./cmd/api

# الواجهة
cd frontend && npm ci
VITE_API_URL=http://127.0.0.1:4000/api npx vite --host 127.0.0.1 --port 5199
# تنفتح على http://127.0.0.1:5199/staff_mange/
```

## الفحوصات (نفسها بالـ CI)
```bash
cd backend-go && go build ./... && go vet ./... && go test -p 1 ./...   # ويا DATABASE_URL تشتغل الاختبارات الحية
cd frontend && npx tsc -p tsconfig.app.json --noEmit                       # ⚠️ مو «tsc --noEmit» لحاله — ذاك ما يفحص شي
cd frontend && npx eslint . && npm test && npm run build
cd frontend && npm run e2e   # Playwright: المالك + ٦ أدوار، يحتاج السيرفر والواجهة شغّالات
```

## الوثائق
**ابدي من هنا**
- [`ARCHITECTURE.md`](ARCHITECTURE.md): البنية العامة، ووين تلگى أي شي.
- [`docs/تعديلات.md`](docs/تعديلات.md): البنود المتفق عليها وحالة كل واحد (المرجع الأساسي للشغل).
- [`docs/luffy-handoff-from-naruto.md`](docs/luffy-handoff-from-naruto.md): سجل التسليم، سطر لكل تعديل.

**الصلاحيات والأمان**
- [`PERMISSIONS.md`](PERMISSIONS.md) و[`docs/الصلاحيات.md`](docs/الصلاحيات.md): كتالوج الصلاحيات ومحرك `internal/authz`.
- [`docs/ماكو-حذف.md`](docs/ماكو-حذف.md): سجل التدقيق واسترجاع المحذوفات.

**التشغيل والطوارئ**
- [`DEPLOY.md`](DEPLOY.md): رفع النظام على السيرفر.
- [`EMERGENCY.md`](EMERGENCY.md): وقت الكارثة، خطوة بخطوة.
- [`RESTORE.md`](RESTORE.md): النسخ الاحتياطي والاسترجاع.
- [`QA-checklist.md`](QA-checklist.md): فحص يدوي حسب كل دور.

**ماتركس والذكاء**
- [`docs/اسلوب-ماتركس.md`](docs/اسلوب-ماتركس.md)، [`docs/محاكاة-ماتركس.md`](docs/محاكاة-ماتركس.md)، [`docs/matrix-chat-plan.md`](docs/matrix-chat-plan.md).
- [`docs/ai-brains-contract.md`](docs/ai-brains-contract.md)، [`docs/secretary.md`](docs/secretary.md)، [`docs/wa-api-contract.md`](docs/wa-api-contract.md).

**تسليمات قديمة بين الوكلاء** (`docs/*-handoff.md`، `docs/sasuke-*`): مرجع تاريخي لشغل مكتمل.

## قواعد ثابتة
- ماكو حذف نهائي بلا موافقة. والحذف ينحفظ بسجل التدقيق ويرجع.
- الرفض يرجّع ٤٠٤ «مو موجود» (`authz.Deny`)، لأن ٤٠٣ ينسجّل مخالفة على الموظف.
- أي عمود جديد بالقاعدة لازم ينضاف للنموذج بـ Go، لأن أغلب القراءات `SELECT *`.
- الراتب وأسماء الموظفين والزبائن ما تطلع لأي ذكاء خارجي.
