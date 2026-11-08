import { useEffect, useState } from 'react'

// (ع) 11-08: عرض مجسّم ثري دي (GLB) جوّه النظام — مكتبة جوجل المفتوحة، تنحمّل بس لما تنفتح.
declare module 'react' {
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace JSX {
    interface IntrinsicElements {
      'model-viewer': React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement> & {
        src?: string; alt?: string; 'camera-controls'?: boolean; 'auto-rotate'?: boolean; ar?: boolean
      }
    }
  }
}

export default function ModelViewer({ src, alt }: { src: string; alt: string }) {
  const [ready, setReady] = useState(false)
  useEffect(() => {
    let alive = true
    import('@google/model-viewer').then(() => { if (alive) setReady(true) }).catch(() => {})
    return () => { alive = false }
  }, [])
  if (!ready) return <p className="p-6 text-center text-xs text-slate-400">جاري تحميل عارض الثري دي…</p>
  return <model-viewer src={src} alt={alt} camera-controls auto-rotate style={{ width: '100%', height: 360, background: '#f1f5f9', borderRadius: 12 }} />
}
