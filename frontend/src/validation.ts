export function validateCustomerName(name: string): string | null {
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length < 4) {
    return 'يجب إدخال الاسم الرباعي الكامل للزبون (4 أسماء على الأقل)'
  }
  return null
}

export function validateCustomerPhone(phone: string): string | null {
  if (!/^\d{11}$/.test(phone.trim())) {
    return 'رقم الهاتف يجب أن يكون 11 رقماً بالضبط'
  }
  return null
}
