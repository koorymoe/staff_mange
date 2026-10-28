# عقد «دماغين للذكاء» — لوفي (خادم) ⇄ ساسكي (واجهة)

قرار دازاي 10-28. الخطة الكاملة: دماغ داخلي (ايريس + ماتركس، مفتاح واحد) ودماغ خارجي (الجودة + الحجوزات، مفتاح واحد)،
ومحادثة حيّة، وسقف من الصفحة، واقتراحات الذكاء.

## ١. المحادثة الحية (جاهز بالخادم على avg)

`POST /api/matrix/chats/{id}/stream` — نفس جسم `/messages`: `{text, image?}`. ADMIN/OWNER بس (مثل الباقي).

- أخطاء التحقق (رسالة فارغة، حد اليوم، الوكيل مطفي) ترجع **JSON عادي** بكود 4xx/5xx قبل البث — نفس `/messages`.
- إذا نجح: `Content-Type: text/event-stream`، وكل حدث سطر `data: {json}` وبعده سطر فارغ:

| type | الحقول | المعنى |
|---|---|---|
| `thinking` | `text` | **قطعة** تفكير (إنكليزي غالباً) — **تنضاف** لما قبلها |
| `tool` | `name` | بدى يفتش بأداة (مثل `booking_summary`) — اعرضها «🔎 checking booking_summary…» |
| `text` | `text` | **الجواب كامل لحد هسه** — بدّل النص كل مرة (مو إضافة). ممكن يرجع يبدي من جديد بعد أداة |
| `done` | `message` | الرد المحفوظ (نفس شكل رد `/messages`: `{id, role, text, steps, createdAt}`) |
| `error` | `text` | فشل بنص البث — اعرض الرسالة ورجّع النص للخانة |

- الطلب POST، فـ`EventSource` ما يصلح — استعمل `fetch` + `res.body.getReader()` وقسّم على `\n\n`.
- التوكن نفس الهيدر الحالي (`Authorization`).
- المسار القديم `/messages` باقي.

**المطلوب بالواجهة (ساسكي) — `MatrixChat.tsx`:**
1. رسالة المستخدم تطلع **فوراً** لحظة الإرسال — حتى لو المحادثة جديدة (`createMatrixChat` يصير بعد عرضها).
2. بدل «ماتركس يفكّر ويفتش…»: صندوق «Thinking» صغير رصاصي (خط صغير، `dir="ltr"`) يتحدث حي بالتفكير وأسطر الأدوات.
   بعد `done` ينطوي تلقائياً لسطر «Thought for Ns · 3 tools» وينفتح بضغطة.
3. الجواب يطلع ويتحدث مع أحداث `text`، وبعد `done` يتبدّل بالرسالة المحفوظة (حتى أزرار التصحيح تشتغل بالـid).
4. زر إيقاف (يلغي الـfetch بـAbortController).

## ٢. ايريس — نفس الشي بعد دمج الدماغين (لوفي يبلّغ بالمسار)

## ٣. السقف (قريباً)
`GET /api/ai/status` يرجّع فوق الموجود:
`brains: {internal: {model, chatModel, spentTodayUSD, dailyLimitUSD}, external: {...}}`
و`PUT /api/ai/budgets/{internal|external}` بجسم `{dailyUSD}` — المالك بس.
**الواجهة:** بصفحة الوكلاء بطاقتين للدماغين: صرف اليوم/السقف بشريط، وخانة تعديل السقف للمالك، والموديل.

## ٤. اقتراحات الذكاء (قريباً)
`GET /api/ai/suggestions?status=` و`POST /api/ai/suggestions/{id}/decide` بجسم `{status: ACCEPTED|REJECTED|DONE}` — OWNER/ADMIN.
`AiSuggestion {id, agent, title, details, needsCode, status, createdAt}`.
**الواجهة:** صفحة «🧠 اقتراحات الذكاء» من مكتب المدير — بطاقات، علامة «يحتاج برمجة»، أزرار، فلتر (الكل / يحتاج برمجة / الجديد).
