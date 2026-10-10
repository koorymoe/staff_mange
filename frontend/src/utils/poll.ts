// ═══ تحديث دوري بس والتبويب ظاهر — (ع) 10-10: «النظام صار ثكيل» ═══
// كل تبويب مفتوح جان يضرب الخادم كل 30–60 ثانية حتى لو مخفي ومنسي.
// هسه: التبويب المخفي ما يطلب شي، ولمن يرجع يظهر يحدّث مرة وحدة إذا
// فاتته دورة. يرجّع دالة إيقاف (بدل clearInterval).
export function pollVisible(fn: () => unknown, ms: number): () => void {
  let last = Date.now()
  const run = () => { last = Date.now(); void fn() }
  const iv = window.setInterval(() => { if (!document.hidden) run() }, ms)
  const onVis = () => { if (!document.hidden && Date.now() - last >= ms) run() }
  document.addEventListener('visibilitychange', onVis)
  return () => { window.clearInterval(iv); document.removeEventListener('visibilitychange', onVis) }
}
