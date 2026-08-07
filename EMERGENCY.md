# 🚨 دليل الطوارئ — نظام شركة الأماني

> اقرا هذا الملف **وقت الكارثة**. مكتوب بالخطوات، انسخ والصق.
> للتفاصيل الكاملة: [`RESTORE.md`](RESTORE.md)

---

## أولاً: شخّص الحالة

| الحالة | روح للقسم |
|---|---|
| النظام ما يفتح / صفحة بيضاء | [١](#١-النظام-ما-يفتح) |
| بيانات ضاعت (حجوزات، زبائن، فواتير) | [٢](#٢-بيانات-ضاعت) |
| السيرفر كله راح | [٣](#٣-السيرفر-كله-راح) |

---

## ١) النظام ما يفتح

```bash
cd ~/staff_mange
docker compose ps          # منو واقف؟
docker compose logs --tail=50 backend
docker compose restart     # أغلب المشاكل تنحل هنا
```

لو ضل واقف:
```bash
docker compose down && docker compose up -d --build
```

**⚠️ قبل أي شي تسويه: خذ نسخة.**
```bash
./backup-db.sh
```

---

## ٢) بيانات ضاعت

```bash
cd ~/staff_mange
ls -lht backups/           # شوف النسخ، الأحدث فوق
./restore-backup.sh        # آخر نسخة

# أو نسخة محددة (مثلاً نسخة أمس قبل ما يصير الخطأ):
./restore-backup.sh backups/staffmange_20260807_030000.tar.gz
```

السكربت:
- يعرضلك شنو جوّا النسخة قبل ما ينفّذ
- **ياخذ نسخة أمان من وضعك الحالي** — لو استرجعت الغلط تكدر ترجع
- يسألك «متأكد؟» ولازم تكتب **نعم**

بعدها:
```bash
docker compose restart backend
```

---

## ٣) السيرفر كله راح

### أ) جيب النسخة

**من R2:**
```bash
aws s3 ls s3://<اسم-الباكت>/backups/ --endpoint-url <R2_ENDPOINT>
aws s3 cp s3://<اسم-الباكت>/backups/staffmange_XXXX.tar.gz.gpg . --endpoint-url <R2_ENDPOINT>
```

**من Hetzner:** لوحة Hetzner → Backups → أنشئ سيرفر من آخر صورة. (تخطى باقي الخطوات، السيرفر يرجع كامل.)

### ب) سيرفر جديد

```bash
# ١) Docker
curl -fsSL https://get.docker.com | sh

# ٢) المشروع
git clone <رابط-المستودع> staff_mange && cd staff_mange

# ٣) النسخة
mkdir -p backups && cp ~/staffmange_XXXX.tar.gz.gpg backups/

# ٤) استخرج .env من النسخة (بدونه ما يشتغل شي)
gpg -d backups/staffmange_XXXX.tar.gz.gpg 2>/dev/null | tar xz -O ./env.txt > .env
chmod 600 .env
cat .env      # تأكد إنه صحيح

# ٥) شغّل
docker compose up -d
sleep 30

# ٦) استرجع
./restore-backup.sh backups/staffmange_XXXX.tar.gz.gpg

# ٧) أعد التشغيل
docker compose restart backend
```

### ج) الدومين

من مزوّد الدومين: وجّه سجل `A` على IP السيرفر الجديد.
Caddy ياخذ شهادة HTTPS تلقائياً خلال دقيقة (أو ترجع من النسخة فوراً).

### د) تأكد

```bash
curl -s https://staffmanage.cc/api/health     # لازم {"status":"ok"}
```
وسجّل دخول بحساب المالك وشوف الحجوزات والصور.

---

## أرقام لازم تعرفها

| السؤال | الجواب |
|---|---|
| وين النسخ؟ | `~/staff_mange/backups/` + R2/المكان الثاني |
| شكد نحتفظ؟ | ١٤ يوم |
| متى تصير؟ | كل يوم ٣:٠٠ فجراً |
| شنو جوّاها؟ | قاعدة البيانات + الملفات + `.env` + إعدادات النشر + الأدلة + شهادات HTTPS |
| مشفّرة؟ | نعم إذا `BACKUP_PASSPHRASE` مضبوطة |
| **وين كلمة سر النسخة؟** | بملف `.env` — **⚠️ احتفظ بنسخة منها بمكان آمن برّا السيرفر، بدونها النسخ ما تنفتح** |

---

## اطمّن قبل ما تحتاجه

```bash
./backup-status.sh
```

يجاوبك: الجدولة شغالة؟ آخر نسخة متى؟ شنو جوّاها؟ اكو نسخة برّا السيرفر؟

**شغّله كل أسبوع.** أخطر شي مو إنك ما عندك نسخ — إنك تظن عندك وهي واقفة من شهرين.

> والنظام يرسللك إشعار داخل التطبيق إذا فشلت النسخة أو ما اكو نسخة خارجية.
