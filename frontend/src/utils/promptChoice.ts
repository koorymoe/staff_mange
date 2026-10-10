// اختيار بسيط برقم من قائمة — لمواضع خفيفة مبنية أصلاً على prompt()/alert()
// خام (نفس أسلوبها) بدل بناء نافذة منبثقة لسطرين استخدام.
export function promptChoice<T extends string>(title: string, options: [T, string][]): T | null {
  const listText = options.map(([, label], i) => `${i + 1}. ${label}`).join('\n')
  const input = prompt(`${title}\n${listText}\n\nاكتب رقم الخيار:`)
  if (input == null || !input.trim()) return null
  const idx = parseInt(input.trim(), 10) - 1
  if (Number.isNaN(idx) || idx < 0 || idx >= options.length) {
    alert('اختيار غير صحيح')
    return null
  }
  return options[idx][0]
}
