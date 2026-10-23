// يطلق «خلص حجز» — LastBookingPrompt يسأل «هذا آخر حجز؟» إذا الدوام خلص.
export function bookingDone() {
  window.dispatchEvent(new Event('matrix-booking-done'))
}
