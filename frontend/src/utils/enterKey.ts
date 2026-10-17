import type { KeyboardEvent } from 'react'

// ═══ زر Enter يشتغل كدور طبيعي بالنظام ═══
//
// «المحاسب يدخل رقم فاتورة، وهو ملزم إنه يضغط (تم). أني أريده عادي
// يضغط Enter ويعتبره تم — أريد الـEnter تشتغل كدور طبيعي بالنظام».
//
// وهذا مو ترف: المحاسب يدخل عشرات الأرقام باليوم، وكل رقم يعني إنه
// يرفع إيده عن لوحة المفاتيح، يمسك الماوس، يدوّر الزر، يضغط، ويرجع.
// خمس حركات بدل وحدة — مضروبة بعشرات المرات باليوم.
//
// ⚠️ ليش دالة مشتركة مو `onKeyDown` مكتوب بكل شاشة؟ لأن التفاصيل
// تنسى: الـ«تركيب» (IME) للكتابة العربية يرسل Enter وهو نص كلمة،
// وShift+Enter لازم يبقى سطراً جديداً بخانات الملاحظات، والزر
// المعطّل ما يجوز ينضغط بالمفتاح وهو مو منضغط بالماوس.

interface EnterOptions {
  /** ما ينفّذ لو صح — نفس شرط تعطيل الزر */
  disabled?: boolean
  /** يسمح Enter داخل خانة نص متعدد الأسطر (الافتراضي: لا) */
  allowInTextarea?: boolean
}

/** يرجّع `onKeyDown` ينفّذ الإجراء عند Enter.
 *
 *  الاستعمال: `<input {...onEnter(() => save())} />`
 */
export function onEnter(action: () => void, options: EnterOptions = {}) {
  return {
    onKeyDown: (e: KeyboardEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
      if (e.key !== 'Enter') return
      // ⚠️ أثناء تركيب الحروف (العربية والفارسية بأنظمة IME) المتصفح
      // يرسل Enter حتى يثبّت الحرف — تنفيذ الحفظ هنا يقطع الكلمة
      // بنصها ويحفظ نصاً ناقصاً.
      if (e.nativeEvent.isComposing) return
      // Shift+Enter يبقى «سطر جديد» — الملاحظات تنكتب بأسطر.
      if (e.shiftKey && !options.allowInTextarea) return
      const el = e.target as HTMLElement
      if (!options.allowInTextarea && el.tagName === 'TEXTAREA') return
      if (options.disabled) return
      e.preventDefault()
      action()
    },
  }
}
