export function openManagerChat(prefill?: string) {
  window.dispatchEvent(new CustomEvent('open-manager-chat', { detail: { prefill } }))
}
