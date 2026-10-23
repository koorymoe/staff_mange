# عقد API الواتساب ووكلاء الذكاء (ناروتو ← ساسكي) — ١٠-٢٣

كل المسارات تحت `/api`. التوقيت ISO بـUTC، والعرض بتوقيت بغداد.
`agent` قيمته `QUALITY` أو `BOOKING`.

**الصلاحيات:**
- كل المسارات: المالك، والمدير (ADMIN)، و`quality_control` لـQUALITY، و`coordinator` لـBOOKING.
- مفاتيح التشغيل والتجربة: **المالك بس**.
- الواجهة ما تنادي مسار مو مسموح للدور، حتى ما يطلع 403.

## المحادثات
### `GET /wa/conversations?agent=QUALITY&status=BOT|HUMAN|CLOSED`
```json
[{
  "id": "uuid", "agent": "QUALITY", "status": "BOT",
  "customerName": "…", "phone": "+9647…", "bookingId": "…", "bookingCode": "B123",
  "sandbox": false, "lastMessageAt": "2026-10-23T16:00:00Z",
  "lastMessage": "نص آخر رسالة (مقطوع ١٢٠ حرف)", "unread": 2
}]
```
- `status`:
  - `BOT`: الوكيل ماسكها.
  - `HUMAN`: إنسان استلمها، والوكيل ساكت.
  - `CLOSED`: خلصت.

### `GET /wa/conversations/{id}`
```json
{
  "conversation": { …نفس الشكل فوق… },
  "messages": [
    { "id": "uuid", "direction": "IN", "author": "CUSTOMER", "text": "…", "createdAt": "…" },
    { "id": "uuid", "direction": "OUT", "author": "AGENT", "text": "…", "createdAt": "…" },
    { "id": "uuid", "direction": "OUT", "author": "HUMAN", "authorName": "اسم الموظف", "text": "…", "createdAt": "…" }
  ],
  "actions": [ …نفس شكل /ai/actions… ]
}
```

### `POST /wa/conversations/{id}/takeover`
بلا جسم. المحادثة تصير `HUMAN`. يرجّع `{ "ok": true }`.

### `POST /wa/conversations/{id}/release`
بلا جسم. ترجع `BOT`. يرجّع `{ "ok": true }`.

### `POST /wa/conversations/{id}/send`
- الجسم: `{ "text": "…" }`، رد إنسان. ينقبل بس إذا الحالة `HUMAN`.
- يرجّع الرسالة بشكل `messages[i]`.

### `POST /wa/conversations/{id}/close`
بلا جسم. يرجّع `{ "ok": true }`.

## إجراءات الوكيل
### `GET /ai/actions?targetId=<bookingId|followUpId>` أو `?conversationId=`
```json
[{
  "id": "uuid", "agent": "QUALITY", "agentName": "ابو الجودة",
  "kind": "RECORD_CONTACT|CREATE_BOOKING|CONFIRM_BOOKING|ADD_DETAILS|SUGGEST_MATRIX|HANDOFF",
  "summary": "سوّى حجز صيانة B130 — الزبون يكول الكاميرا تطفي",
  "reason": "…", "targetType": "BOOKING|FOLLOWUP|PROPOSAL", "targetId": "…",
  "sandbox": false, "undoable": true,
  "undoneAt": null, "undoneByName": null, "createdAt": "…"
}]
```

### `POST /ai/actions/{id}/undo`
- الجسم: `{ "reason": "…" }`.
- يرجّع الإجراء بعد التحديث.
- إلغاء `CREATE_BOOKING` **يرفع طلب حذف عادي** للحجز، مو حذف مباشر.

## مفاتيح التشغيل (المالك)
### `GET /ai/agents`
موجود. وينضاف لكل وكيل `"enabled": bool` و`"built": true|false`.

### `PUT /ai/agents/{QUALITY|BOOKING}/enabled`
- الجسم: `{ "enabled": true }`.
- يرجّع `{ "agent": "QUALITY", "enabled": true }`.
- إذا المفتاح مشتغل بس الواتساب مو مربوط بعد، يرجّع 409 ويا الرسالة «الواتساب مو مربوط بعد — شغّل وضع التجربة».

## وضع التجربة (المالك)
### `POST /wa/sandbox/start`
- الجسم: `{ "agent": "QUALITY", "bookingId": "…" }`.
- يرجّع `{ "conversation": {…}, "messages": [ …رسالة الترحيب… ] }`.

### `POST /wa/sandbox/{id}/say`
- الجسم: `{ "text": "…" }`، كأنه الزبون.
- يرجّع `{ "messages": [ …رسالة الزبون + رد الوكيل… ], "actions": [ …إجراءات هذا الدور… ] }`.
- إجراءات التجربة `sandbox: true`، و**ما تنفّذ على النظام**.

## الحالة
`GET /ai/status` موجود. وينضاف `quality` و`booking` بنفس شكل `matrix`، و`whatsapp: { "on": bool }`.
