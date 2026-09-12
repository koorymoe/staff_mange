// ═══ الوضع الليلي — التطبيق والحفظ ═══
//
// بملف منفصل عن الزر حتى يضل التحديث السريع (Fast Refresh) شغّال
// بالتطوير، ولأن `main.tsx` يحتاجها **قبل** ما يوجد أي مكوّن.

const KEY = 'theme'

/** يطبّق الوضع على `<html>`. */
export function applyTheme(dark: boolean) {
  document.documentElement.classList.toggle('dark', dark)
}

/** ⚠️ الاختيار ينحفظ **بالجهاز** مو بحساب الموظف: نفس الموظف يشتغل
 *  بالمكتب على حاسبة بغرفة مضوية، وبالليل من تلفونه بالبيت — وربطه
 *  بالحساب يعني إنه يبدّله مرتين باليوم.
 *
 *  وأول مرة نتبع إعداد الجهاز نفسه: الي محوّل تلفونه للوضع الليلي
 *  يتوقع كل شي يفتحه يكون ليلياً. */
export function prefersDark(): boolean {
  const saved = localStorage.getItem(KEY)
  if (saved) return saved === 'dark'
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ?? false
}

export function saveTheme(dark: boolean) {
  localStorage.setItem(KEY, dark ? 'dark' : 'light')
}
