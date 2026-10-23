import { apiRequest } from '../api'

/** (ع 10-16) طلب تفسير مباشر من خانة «المتأخر» — نفس مسار «التبريرات» */
export async function askJustification(employeeId: string, question: string) {
  await apiRequest('/justifications/direct', { method: 'POST', body: JSON.stringify({ employeeId, question }) })
  alert('انطلب التفسير — يطلعله أول ما يفتح النظام، وجوابه يوصلك بـ«التبريرات» بالصندوق')
}
