import { useEffect, useState } from 'react'
import { applyTheme, prefersDark, saveTheme } from '../utils/theme'

// ═══ زر الوضع الليلي ═══
//
// «سوي الوضع الليلي».
//
// ⚠️ الاختيار ينحفظ **بالجهاز** مو بحساب الموظف: نفس الموظف يشتغل
// بالمكتب على حاسبة بغرفة مضوية، وبالليل من تلفونه بالبيت. ربط
// الاختيار بالحساب يعني إنه يبدّله مرتين باليوم.
//
// ⚠️ وأول مرة نتبع إعداد الجهاز نفسه (`prefers-color-scheme`): الي
// محوّل تلفونه للوضع الليلي يتوقع كل شي يفتحه يكون ليلياً.

// التطبيق والحفظ بملف `utils/theme.ts` — يحتاجهن `main.tsx` قبل ما
// يوجد أي مكوّن، ويخلّي التحديث السريع بالتطوير شغّالاً.

export default function ThemeToggle() {
  const [dark, setDark] = useState(prefersDark)

  useEffect(() => { applyTheme(dark) }, [dark])

  return (
    <button
      onClick={() => {
        const next = !dark
        setDark(next)
        saveTheme(next)
      }}
      className="flex h-10 w-10 items-center justify-center rounded-xl text-slate-400 transition-all hover:bg-slate-100 hover:text-slate-600"
      title={dark ? 'رجّع الوضع النهاري' : 'الوضع الليلي — أريح للعين بالليل'}
      aria-label={dark ? 'الوضع النهاري' : 'الوضع الليلي'}
    >
      {dark ? (
        // شمس
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
          <circle cx="12" cy="12" r="4" />
          <path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
        </svg>
      ) : (
        // هلال
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" />
        </svg>
      )}
    </button>
  )
}
