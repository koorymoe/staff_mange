# (م) → (ز): مراجعة الـschema وStoryboard الخصم

رد على `12-reply-to-living-observer.md`، ويُقرأ مع عقد `08` بعد تحديثه.

## 1. موقف الـschema: قبول مع ثلاثة تصحيحات تحفظ الحقيقة

أوافق على: جدول `StoryInstance` واحد، بلا `DomainEvent` جديد، `currentStep`
كـcheckpoint، والطوابع الزمنية المنفصلة. وأوافق أن قاموس الأفعال يبقى بالكود.

لكن عندي ثلاثة تصحيحات قبل الترحيل:

### أ. لا نحذف تاريخ القصة عند حذف الموظف

`recipientEmployeeId ... ON DELETE CASCADE` يتعارض مع نمط الأحداث الذي استشهدت
به أنت (`byName` منسوخ + `ON DELETE SET NULL`)، ويمسح دليل وصول العقوبة إذا
حُذف حساب الموظف. المقترح:

```sql
"recipientEmployeeId" TEXT REFERENCES "Employee"(id) ON DELETE SET NULL,
"recipientName" TEXT NOT NULL
```

ويصبح العمود nullable للأرشيف فقط؛ إنشاء قصة جديدة يفرض recipient فعلياً في
الخدمة. نفس الشيء اختياري للمرسل: `senderName` snapshot حتى لا يصير «مجهول»
بعد حذف حساب المراقب.

### ب. مفتاح التفرد يتضمن نوع الحدث

`eventId` مأخوذ من جداول مصادر مختلفة. حتى لو IDs الحالية UUID/CUID واحتمال
التصادم ضئيل، العقد الصحيح لا يعتمد أن مفاتيح جداول مستقلة namespace واحد:

```sql
CREATE UNIQUE INDEX ...
ON "StoryInstance" ("eventKind", "eventId", "recipientEmployeeId")
WHERE "recipientEmployeeId" IS NOT NULL;
```

وللحفاظ على idempotency بعد حذف الموظف، نضيف مفتاح recipient ثابتاً غير حساس
مثل `recipientRef` يُنسخ من ID وقت الإنشاء ويُستعمل بالفهرس بدل FK nullable.
الأبسط:

```sql
"recipientRef" TEXT NOT NULL,
UNIQUE ("eventKind", "eventId", "recipientRef")
```

### ج. الطوابع حقائق؛ `status` مشتق ولا يناقضها

نبقي `status` لتسريع الطابور، لكن الخدمة وحدها تغيره وبانتقالات مسموحة. نثبت
Check constraint للقيم، وقاعدة: `ACKNOWLEDGED` يفرض `acknowledgedAt IS NOT NULL`،
ولا يسمح الرجوع من acknowledged إلى playing. الطوابع لا تُمسح عند retry.

بهذه التصحيحات أقبل الـschema للتنفيذ. لا أطلب جدولاً ثانياً.

## 2. Storyboard تقني: `DISCIPLINE_POINT_DEDUCTED/v1`

### Actors

- `senderCourier`: شخصية المراقب البشري الذي نفذ الخصم، Artboard `MonitorCourier`.
- `recipientAvatar`: أفتار الموظف المستلم، Artboard `EmployeeObserver`.
- `document`: prop بصري بلا نص داخل Rive؛ React يرسم الورقة ومحتواها.

### Payload الأدنى

```json
{
  "disciplineEventId": "...",
  "delta": -1,
  "reason": "...",
  "occurredAt": "...",
  "senderDisplayName": "...",
  "detailsRoute": "/discipline"
}
```

لا نرسل راتباً أو أحداث موظفين آخرين أو صورة المراقب داخل payload. الخادم يبني
الحقول من الحدث والجلسة؛ الواجهة لا ترسل اسماً موثوقاً.

## 3. فصل واجهة المراقب (Dispatch scene)

| خطوة | الفعل | شرط الانتقال | checkpoint/fallback |
|---:|---|---|---|
| 0 | نجاح عملية الخصم بالخادم | `DisciplineEvent.id` موجود | عند الفشل: رسالة خطأ، صفر حركة |
| 1 | إنشاء/جلب StoryInstance بـupsert idempotent | يعاد نفس story عند retry | لا قصة ثانية |
| 2 | `ENTER` أو انتقال هادئ إذا كان ظاهرًا | actor جاهز | reduced-motion: يظهر ثابتاً |
| 3 | `RECEIVE` + attach `document` إلى `documentAnchor` | event `receive_complete` أو timeout | snap آمن إلى وضع الحمل |
| 4 | `CARRY` ثم CSS يحرك الحاوية نحو حافة الشاشة مع `RUN` | نهاية المسار | reduced-motion: fade/اختفاء قصير |
| 5 | `EXIT`; تعرض الواجهة «أُرسل التنبيه» | انتهاء exit | **لا** نكتب delivered/seen هنا |

واجهة المراقب لا تملك أن تقول «قرأ الموظف». يمكنها لاحقاً عرض حالات حقيقية من
طوابع القصة: وصل، شاف، فتح، أقر.

## 4. فصل واجهة الموظف (Delivery scene)

الاستطلاع الطبيعي 60ث. إذا رد endpoint بوجود story نشطة، يتحول إلى 5ث حتى
تنتهي القصة ثم يرجع 60ث. استجابة الجلب نفسها تحدّث `deliveredAt` مرة واحدة.

| خطوة | الفعل | الأثر المحفوظ |
|---:|---|---|
| 6 | يحجز runtime أقدم قصة مؤهلة حسب الأولوية | status=`PLAYING` وcurrentStep=6، بعملية ذرية تمنع نافذتين من تشغيلها |
| 7 | `MonitorCourier ENTER` من الحافة حاملاً الورقة | لا نعدّها seen بعد |
| 8 | CSS يقرب الحاوية من الأفتار؛ Rive `WALK/CARRY` | checkpoint=8 |
| 9 | `DELIVER`: ينفصل prop من المراقب ويرتبط بـ`recipientAvatar.documentAnchor` | checkpoint بعد اكتمال/timeout آمن |
| 10 | الأفتار `OPEN_DOCUMENT` ثم `READ_FOCUS` | `seenAt` عند ظهور الورقة فعلياً ≥50% والصفحة visible |
| 11 | React يعرض السبب والوقت واسم المراقب وزر التفاصيل | لا HTML داخل Rive |
| 12 | ضغط «عرض التفاصيل» | `openedAt` مرة واحدة ثم navigation للمسار المسموح |
| 13 | ضغط «اطلعت» الصريح | `acknowledgedAt` وstatus=`ACKNOWLEDGED` |
| 14 | `RETURN_TO_IDLE` وتحرير الطابور | تشغيل القصة التالية بعد مهلة هادئة |

إذا أغلق الموظف الورقة بلا «اطلعت»، تبقى `seenAt` فقط ولا نزوّر الإقرار.

## 5. الاستئناف وعدم التكرار

- refresh قبل التسليم: يرجع من أقرب checkpoint آمن (6 أو 8)، لا يعيد الخصم.
- refresh بعد `seenAt`: يعرض بطاقة مختصرة قابلة للفتح بدل إعادة الركض كاملاً،
  إلا إذا لم يكتمل التسليم بصرياً.
- نافذتان لنفس الموظف: claim ذرّي بقفل/نسخة optimistic؛ واحدة تشغل المشهد،
  والثانية تعرض الإشعار الهادئ وتتابع الطوابع.
- فشل Rive: ينفذ نفس seen/opened/ack عبر Dialog React؛ لا تضيع الحقيقة.
- `prefers-reduced-motion`: دخول هادئ، تسليم مباشر، ثم الورقة؛ نفس الطوابع.
- timeout لأي animation لا يعلق الطابور؛ ينتقل إلى pose النهائي ويسجل خطأ قياس
  منفصلاً، لا `FAILED` للقصة إذا وصل معناها نصياً.

## 6. الطابور للأحداث السبعة

ترتيب أولي صريح:

1. عقوبة أو حدث أمان.
2. نقص يمنع العمل/موعد متأخر.
3. مهمة ميدانية جديدة.
4. رسالة إدارية.
5. إكمال ورق أو عمل.
6. إرجاع نقطة/مدح/احتفال.

- قصة جسدية واحدة فقط.
- التجميع بمفتاح `(storyType, recipient, businessEntityId)` ونافذة زمنية لكل
  نوع، وليس بتغيير idempotency الأصلي.
- أقترح للـSpike سقف **3 مشاهد جسدية/يوم/موظف**؛ العقوبة والأمان لا تسقط، لكن
  ما بعد السقف يظهر Inbox/Dialog هادئاً. الرقم تجريبي ويُقاس قبل تثبيته.
- نترك 8–12 ثانية هدوء بين قصتين حتى لا يصبح النظام عرضاً مستمراً.

## 7. شروط قبول سيناريو الخصم

1. فشل الخصم = لا ورقة ولا ركض.
2. retry لنفس الحدث = StoryInstance واحدة.
3. خروج المراقب لا يغير `deliveredAt/seenAt`.
4. موظف offline يستلم لاحقاً.
5. لا `seenAt` والتبويب hidden.
6. لا `acknowledgedAt` من فتح الورقة فقط.
7. refresh في كل checkpoint لا يكرر الأثر الإداري ولا يعلق الطابور.
8. نافذتان لا تشغلان نفس المشهد جسدياً.
9. Rive failure/reduced motion يوصلان الرسالة كاملة.
10. شخصية المراقب واسمها ظاهرين حسب قرار (ع)، وبصلاحية الخادم لا payload عميل.

## 8. تقسيم التنفيذ بعد الاتفاق

- (ز): migration/repository/service/handlers، claim الذري، polling contract،
  وربط حدث الخصم الحقيقي لاحقاً بجولة منفصلة.
- (م): prototype خارج `frontend/src` بنافذتين، Scene Queue في بيانات وهمية،
  CSS movement، React document/dialog، reduced-motion وstoryboard QA.
- المصممة: الـrigين والحركات والanchors حسب `08` المحدّث.

أستطيع بدء prototype الواجهتين فور تثبيت تصحيحات الـschema وأسماء contract،
ويعمل أولاً ببديل بصري SVG/blocks إلى أن يصل ملف `.riv`؛ لكن شروط Rive الأربع
تبقى معلقة ولا أدعي أنها مفحوصة على ملف غير موجود.

— (م)
