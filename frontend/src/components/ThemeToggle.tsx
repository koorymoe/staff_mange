import { useEffect, useState } from 'react'
import { applyTheme, getTheme, saveTheme, type Theme } from '../utils/theme'

// ═══ زر الثيم — نهاري / ليلي / عين الأماني ═══
//
// «سوي الوضع الليلي» صار ثلاثة خيارات بعد إضافة «عين الأماني»
// (الطابع التقني الكحلي-السماوي). البنية segmented control صغيرة
// بنفس حجم ومكان الزر الثنائي القديم — بلا شاشة جديدة.
//
// ⚠️ الاختيار ينحفظ **بالجهاز** مو بحساب الموظف: نفس الموظف يشتغل
// بالمكتب على حاسبة بغرفة مضوية، وبالليل من تلفونه بالبيت. ربط
// الاختيار بالحساب يعني إنه يبدّله مرتين باليوم.
//
// التطبيق والحفظ بملف `utils/theme.ts` — يحتاجهن `main.tsx` قبل ما
// يوجد أي مكوّن، ويخلّي التحديث السريع بالتطوير شغّالاً.

const OPTIONS: { value: Theme; label: string; title: string }[] = [
  { value: 'light', label: '☀️', title: 'الوضع النهاري' },
  { value: 'dark', label: '🌙', title: 'الوضع الليلي — أريح للعين بالليل' },
  { value: 'amani-eye', label: '👁️', title: 'عين الأماني' },
]

export default function ThemeToggle() {
  const [theme, setTheme] = useState<Theme>(getTheme)

  useEffect(() => { applyTheme(theme) }, [theme])

  return (
    <div
      className="flex items-center gap-0.5 rounded-xl bg-slate-100 p-1 dark:bg-slate-800"
      role="radiogroup"
      aria-label="ثيم النظام"
    >
      {OPTIONS.map((opt) => (
        <button
          key={opt.value}
          onClick={() => { setTheme(opt.value); saveTheme(opt.value) }}
          className={`flex h-8 w-8 items-center justify-center rounded-lg text-sm transition-all ${
            theme === opt.value
              ? 'bg-white shadow-sm dark:bg-slate-600'
              : 'opacity-50 hover:opacity-80'
          }`}
          title={opt.title}
          aria-label={opt.title}
          aria-pressed={theme === opt.value}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}
