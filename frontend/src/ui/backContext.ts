import { createContext } from 'react'

// true داخل الـLayout: زر «→ رجوع» العام موجود فوگ كل شاشة، فأزرار الرجوع داخل الصفحات تختفي حتى ما يتكرر.
export const GlobalBackContext = createContext(false)
