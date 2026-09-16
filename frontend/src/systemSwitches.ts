import { api } from './api'

// ═══ مفاتيح إطفاء الميزات — قراءة الواجهة ═══
//
// 🔴 **ليش وحدة لحالها**: مكوّنان منفصلان (الكائن والإعلانات) يحتاجان
// نفس الجواب، وكل واحد ينادي لحاله يعني **نداءان بكل تنقّل لكل
// موظف** — ونفس الجواب.
//
// ⚠️ **والافتراضي «شغّال» لمّا نفشل**: لو رد الخادم خطأ أو الشبكة
// مقطوعة، نفترض إن الميزات شغّالة. والعكس يعني إن خللاً بالشبكة
// **يطفّي نظام الموظف** بهدوء — وهذا أسوأ بكثير من إعلان يبين لحظة
// زيادة.

export const SWITCH_ENTITY = 'entity_enabled'
export const SWITCH_ANNOUNCEMENTS = 'announcements_enabled'

type Switches = Record<string, boolean>

let cache: Promise<Switches> | null = null

export function getSwitches(): Promise<Switches> {
  if (!cache) {
    cache = api.getSystemSwitches().catch(() => ({} as Switches))
  }
  return cache
}

/** ⚠️ **تُنادى بعد أي تبديل**: بلاها (ع) يطفّي المفتاح وما يتغيّر
 *  شي على شاشته هو لحد ما يعيد تحميل الصفحة، فيحسب إنه ما انحفظ. */
export function forgetSwitches(): void {
  cache = null
}

/** هل الميزة شغّالة؟ المجهول = شغّالة (الميزة الي انبنت تبقى تشتغل). */
export async function isEnabled(key: string): Promise<boolean> {
  const all = await getSwitches()
  return all[key] !== false
}
