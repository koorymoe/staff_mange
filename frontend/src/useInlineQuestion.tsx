import { useCallback, useEffect, useRef, useState } from 'react'
import InlineQuestion from './components/InlineQuestion'
export function useInlineQuestion() {
  const [question, setQuestion] = useState<{
    message: string
    reasonRequired: boolean
  } | null>(null)
  const resolver = useRef<((result: string | null) => void) | null>(null)
  useEffect(
    () => () => {
      resolver.current?.(null)
      resolver.current = null
    },
    [],
  )
  const ask = (
    message: string,
    reasonRequired = false,
  ): Promise<string | null> => {
    if (resolver.current) return Promise.resolve(null)
    return new Promise((resolve) => {
      resolver.current = resolve
      setQuestion({ message, reasonRequired })
    })
  }
  const finish = useCallback((result: string | null) => {
    const resolve = resolver.current
    resolver.current = null
    setQuestion(null)
    resolve?.(result)
  }, [])
  return {
    ask,
    dialog: question ? <InlineQuestion {...question} finish={finish} /> : null,
  }
}
