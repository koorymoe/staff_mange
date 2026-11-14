import { useEffect, useRef } from 'react'

// (دازاي 11-15) «فتحت وحدة ينقص واحد»: الحجز الموجّه ينحسب «انفتح» من يبقى قدام عين الموظف
// ثانية ونص (مو أول ما الصفحة تنفتح كلها) — فالرقم الأحمر ينقص وحدة وحدة ويا كل حجز يشوفه.
const sent = new Set<string>()

export function useSeenOnView<T extends HTMLElement>(id: string, mark: (id: string) => Promise<unknown>) {
  const ref = useRef<T | null>(null)
  useEffect(() => {
    const el = ref.current
    if (!el || sent.has(id) || typeof IntersectionObserver === 'undefined') return
    let timer: ReturnType<typeof setTimeout> | undefined
    const io = new IntersectionObserver(([e]) => {
      if (e.isIntersecting) {
        timer = setTimeout(() => {
          if (sent.has(id)) return
          sent.add(id)
          mark(id).then(() => window.dispatchEvent(new Event('matrix-courier-refresh'))).catch(() => sent.delete(id))
          io.disconnect()
        }, 1500)
      } else if (timer) clearTimeout(timer)
    }, { threshold: 0.6 })
    io.observe(el)
    return () => { io.disconnect(); if (timer) clearTimeout(timer) }
  }, [id, mark])
  return ref
}
