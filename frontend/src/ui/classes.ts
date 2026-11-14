// ═══ (DRY 11-02) ستايلات مشتركة ═══
// الكارت الأبيض المعتاد — چان مكتوب حرفياً بأربع شاشات.
export const card = 'rounded-2xl border border-slate-200 bg-white p-4 shadow-sm'

// ═══ (دازاي 11-15) ألوان الأزرار الموحّدة ═══
// الزر الرئيسي كحلي وحد بكل النظام — چان سبع ألوان (brand/أزرق/نيلي/بنفسجي…).
// الأخضر للموافقة، الأحمر للحذف/الخطر، والخفيف للإلغاء والثانوي.
const BTN_BASE = 'inline-flex items-center justify-center gap-1.5 rounded-lg px-4 py-2 text-sm font-bold transition disabled:cursor-not-allowed disabled:opacity-50'
export const BTN_PRIMARY = `${BTN_BASE} bg-[#0f2040] text-white hover:bg-[#1a3a6c]`
export const BTN_DANGER = `${BTN_BASE} bg-red-600 text-white hover:bg-red-700`
export const BTN_SUCCESS = `${BTN_BASE} bg-emerald-600 text-white hover:bg-emerald-700`
export const BTN_GHOST = `${BTN_BASE} border border-slate-200 bg-white text-slate-700 hover:bg-slate-50`
