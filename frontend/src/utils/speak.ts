import { api } from '../api'

// 🔊 (دازاي 11-17) تشغيل صوت رد كمال — واحد بالمرة.
let current: HTMLAudioElement | null = null

export async function speakText(text: string): Promise<void> {
  const blob = await api.secretarySpeak(text)
  const url = URL.createObjectURL(blob)
  current?.pause()
  const a = new Audio(url)
  current = a
  a.onended = () => URL.revokeObjectURL(url)
  await a.play()
}

export function stopSpeaking() {
  current?.pause()
  current = null
}
