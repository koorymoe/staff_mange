// ═══ ثيمات النظام — التطبيق والحفظ ═══
//
// بملف منفصل عن الزر حتى يضل التحديث السريع (Fast Refresh) شغّال
// بالتطوير، ولأن `main.tsx` يحتاجها **قبل** ما يوجد أي مكوّن.

export type Theme = 'light' | 'dark' | 'amani-eye'

const KEY = 'theme'

/** يطبّق الثيم على `<html>`.
 *
 * ⚠️ «عين الأماني» غامقة دايماً — نبقي صنف `dark` معاها حتى كل
 * تجاوز `html.dark` موجود بالكود يشتغل تلقائياً (خانات إدخال،
 * ظلال...)، وفوقه تجاوزات `theme-amani-eye` الخاصة تاخذ الأولوية
 * بترتيب الكتابة بـ`index.css`. */
export function applyTheme(theme: Theme) {
  const root = document.documentElement
  root.classList.toggle('dark', theme === 'dark' || theme === 'amani-eye')
  root.classList.toggle('theme-amani-eye', theme === 'amani-eye')
}

/** ⚠️ الاختيار ينحفظ **بالجهاز** مو بحساب الموظف: نفس الموظف يشتغل
 *  بالمكتب على حاسبة بغرفة مضوية، وبالليل من تلفونه بالبيت — وربطه
 *  بالحساب يعني إنه يبدّله مرتين باليوم.
 *
 *  وأول مرة نتبع إعداد الجهاز نفسه للّيلي/النهاري بس: الي محوّل
 *  تلفونه للوضع الليلي يتوقع كل شي يفتحه يكون ليلياً. عين الأماني
 *  خيار صريح بس — ماكو إعداد جهاز يفعّلها تلقائياً. */
export function getTheme(): Theme {
  const saved = localStorage.getItem(KEY)
  if (saved === 'amani-eye' || saved === 'dark' || saved === 'light') return saved
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function saveTheme(theme: Theme) {
  localStorage.setItem(KEY, theme)
}
