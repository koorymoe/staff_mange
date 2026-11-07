import { describe, expect, it } from 'vitest'
import { UR } from './ur'

// بلاغ #11: كل نص ظاهر بشاشات الفني إله ترجمة أوردو — حتى ما يضيف أحد نص جديد وينسى يترجمه.
const FILES = ['pages/MyTasks.tsx', 'pages/MyInventory.tsx', 'components/WorkGate.tsx', 'components/ReportProblemButton.tsx', 'components/WorkPhotoButton.tsx', 'components/RoutePlanCard.tsx']
const SRC = import.meta.glob(['../pages/MyTasks.tsx','../pages/MyInventory.tsx','../components/WorkGate.tsx','../components/ReportProblemButton.tsx','../components/WorkPhotoButton.tsx','../components/RoutePlanCard.tsx'], { query: '?raw', import: 'default', eager: true }) as Record<string, string>
const strip = (s: string) => s.replace(/\{\/\*[\s\S]*?\*\/\}/g, '').replace(/\/\*[\s\S]*?\*\//g, '').replace(/(^|[^:])\/\/.*/g, '$1')
const ar = /[؀-ۿ]/

describe('قاموس الأوردو', () => {
  it('يغطي النصوص الثابتة بشاشات الفني', () => {
    const missing: string[] = []
    for (const f of FILES) {
      const s = strip(SRC['../' + f] ?? '')
      for (const m of s.matchAll(/>([^<>{}]{1,60})[<{]/g)) {
        const t = m[1].split(/\s+/).join(' ').trim()
        if (t && ar.test(t) && t.length > 1 && !/[`$=;]/.test(t) && !UR[t]) missing.push(`${f}: ${t}`)
      }
      for (const m of s.matchAll(/(?:placeholder|title|aria-label)="([^"]+)"/g)) if (ar.test(m[1]) && !UR[m[1]]) missing.push(`${f}: ${m[1]}`)
    }
    expect(Object.keys(SRC).length).toBe(FILES.length)
    expect(missing).toEqual([])
  })
  it('ماكو ترجمة فاضية', () => {
    expect(Object.entries(UR).filter(([, v]) => !v.trim())).toEqual([])
  })
})
