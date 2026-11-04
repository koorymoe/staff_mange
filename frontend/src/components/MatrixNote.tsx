import type { ReactNode } from 'react'

// تعليق أو توجيه من ماتركس — نفس اللون بكل مكان بالنظام (طلب (ع) 10-04).
export default function MatrixNote({ children, className = '', tag = true }: { children: ReactNode; className?: string; tag?: boolean }) {
  return (
    <div className={`matrix-note ${className}`}>
      {tag && <>🤖 <b className="mx-tag">ماتركس:</b> </>}
      {children}
    </div>
  )
}
