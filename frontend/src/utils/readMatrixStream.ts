import type { MatrixChatMessage } from '../api'

export type MatrixStreamEvent =
  | { type: 'thinking' | 'text'; text: string }
  | { type: 'tool'; name: string }
  | { type: 'done'; message: MatrixChatMessage }

// SSE boundaries can split across byte chunks, including inside Arabic UTF-8.
export async function readMatrixStream(
  body: ReadableStream<Uint8Array>,
  emit: (event: MatrixStreamEvent) => void,
  signal: AbortSignal,
) {
  const reader = body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let done = false
  const frame = (raw: string) => {
    const data = raw
      .split(/\r?\n/)
      .filter((line) => line.startsWith('data:'))
      .map((line) => line.slice(5).replace(/^ /, ''))
      .join('\n')
    if (!data) return
    const event = JSON.parse(data)
    if (event.type === 'error') throw new Error(event.text || 'تعذر إكمال الرد')
    if (event.type === 'done') {
      if (
        !event.message?.id ||
        event.message.role !== 'ASSISTANT' ||
        typeof event.message.text !== 'string'
      )
        throw new Error('وصل جواب غير مكتمل من الخادم')
      done = true
      emit(event)
    } else if (
      (event.type === 'thinking' || event.type === 'text') &&
      typeof event.text === 'string'
    )
      emit(event)
    else if (event.type === 'tool' && typeof event.name === 'string')
      emit(event)
  }
  try {
    while (!done) {
      if (signal.aborted) throw new DOMException('Aborted', 'AbortError')
      const chunk = await reader.read()
      buffer += decoder.decode(chunk.value, { stream: !chunk.done })
      let boundary: RegExpExecArray | null
      while (!done && (boundary = /\r?\n\r?\n/.exec(buffer))) {
        const raw = buffer.slice(0, boundary.index)
        buffer = buffer.slice(boundary.index + boundary[0].length)
        frame(raw)
      }
      if (chunk.done) {
        if (buffer.trim() && !done) frame(buffer)
        break
      }
    }
    if (!done) throw new Error('انقطع البث قبل حفظ الجواب. نصك رجع للخانة.')
  } finally {
    await reader.cancel().catch(() => {})
    reader.releaseLock()
  }
}
