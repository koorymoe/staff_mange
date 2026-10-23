// نفس SecretaryHonorific بالخادم؛ ما يعتمد على الاسم أو الدور.
export function irisHonorific(certificate?: string | null): string {
  const c = certificate?.toLowerCase() || ''
  if (
    ['دكتوراه', 'دكتوراة', 'دكتوراء', 'دكتور', 'phd', 'ph.d', 'doctor'].some(
      (x) => c.includes(x),
    )
  )
    return 'دكتور'
  if (['ماجستير', 'master', 'msc', 'm.sc'].some((x) => c.includes(x)))
    return 'أستاذ'
  if (['هندسة', 'مهندس', 'engineer', 'b.eng'].some((x) => c.includes(x)))
    return 'مهندس'
  return 'سيدنا'
}
