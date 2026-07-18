# رفع النظام على سيرفر Hetzner — خطوة بخطوة

## 1. إنشاء السيرفر
- Hetzner Cloud Console → Add Server
- الموقع: أقرب مكان (مثلاً Falkenstein/Nuremberg ألمانيا أو Ashburn أمريكا)
- الصورة (Image): **Ubuntu 24.04**
- النوع: **CX22**
- SSH Key: أضف مفتاح SSH (أو استخدم كلمة سر مؤقتة تظهر بالإيميل)
- اضغط Create & Buy Now

## 2. الاتصال بالسيرفر
```
ssh root@<IP السيرفر>
```

## 3. تثبيت Docker (مرة وحدة بس)
```bash
curl -fsSL https://get.docker.com | sh
```

## 4. تحميل المشروع
```bash
git clone <رابط الريبو> staff_mange
cd staff_mange
git checkout claude/zen-mendel-70b9j2
```

## 5. إعداد الإعدادات
```bash
cp .env.production.example .env
nano .env   # عبّي DB_PASSWORD و JWT_SECRET و CORS_ORIGIN
```
- `JWT_SECRET`: نص عشوائي طويل — تكدر تولّده بـ `openssl rand -hex 32`
- `CORS_ORIGIN`: رابط الموقع (مثلاً `http://<IP السيرفر>` مبدئياً، وبعدين الدومين لما يجهز)

## 6. التشغيل
```bash
./deploy.sh
```
أول مرة تاخذ بضع دقائق (تحميل وبناء الصور). بعدها النظام يشتغل على:
- الموقع: `http://<IP السيرفر>` (بورت 80)
- الباك إند وراء نفس الموقع تلقائياً على `/api`

## 7. كل تحديث لاحق (بعد ما تسوي تعديلات وتدزها Github)
على السيرفر فقط:
```bash
cd staff_mange
./deploy.sh
```

## 8. ربط دومين (اختياري، بعدين)
لما ياخذ الدومين، وجّه سجل A بالدومين لـ IP السيرفر، وبعدها نضيف HTTPS مجاني (Let's Encrypt).

## ملاحظات أمان
- ما تحط ملف `.env` بالـ git أبداً (موجود بالفعل ضمن `.gitignore`).
- خلي `JWT_SECRET` و `DB_PASSWORD` قويين وما تشاركهم مع أحد.
