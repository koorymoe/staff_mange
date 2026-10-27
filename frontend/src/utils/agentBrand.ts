import { apiRequest } from '../api'

// ═══ أسماء الوكلاء (دازاي 10-27) ═══
// «ماتركس» و«ايريس» مكتوبة بمئات الأماكن. بدل ما نلمسها كلها، إذا المالك غيّر الاسم
// نبدّل النص الظاهر بالشاشة (مو الحقول) — وإذا الاسم نفسه الأصلي ما يشتغل شي.

const DEFAULTS: [RegExp, 'MATRIX' | 'IRIS'][] = [
  [/ماتركس/g, 'MATRIX'],
  [/[اإ]يريس/g, 'IRIS'],
]
let names: Record<string, string> = {}
let observer: MutationObserver | null = null

function fix(text: string) {
  let out = text
  for (const [re, key] of DEFAULTS) if (names[key]) out = out.replace(re, names[key])
  return out
}

function walk(root: Node) {
  if (root.nodeType === Node.TEXT_NODE) {
    const v = root.nodeValue ?? ''
    if (/ماتركس|[اإ]يريس/.test(v)) {
      const n = fix(v)
      if (n !== v) root.nodeValue = n
    }
    return
  }
  if (root instanceof HTMLElement && (root.isContentEditable || root.tagName === 'TEXTAREA' || root.tagName === 'SCRIPT' || root.tagName === 'STYLE')) return
  root.childNodes.forEach(walk)
}

export function startAgentRebrand() {
  apiRequest<Record<string, string>>('/ai/agent-names')
    .then((n) => {
      names = {}
      if (n.MATRIX && n.MATRIX !== 'ماتركس') names.MATRIX = n.MATRIX
      if (n.IRIS && n.IRIS !== 'ايريس') names.IRIS = n.IRIS
      observer?.disconnect()
      observer = null
      if (!names.MATRIX && !names.IRIS) return
      walk(document.body)
      document.title = fix(document.title)
      observer = new MutationObserver((ms) => {
        for (const m of ms) {
          if (m.type === 'characterData') walk(m.target)
          else m.addedNodes.forEach(walk)
        }
      })
      observer.observe(document.body, { subtree: true, childList: true, characterData: true })
    })
    .catch(() => {})
}
