import { apiRequest } from '../api'

import { notify } from '../ui/notify'
/** (ع 10-16) طلب تفسير مباشر من خانة «المتأخر» — نفس مسار «التبريرات» */
export async function askJustification(employeeId: string, question: string) {
  await apiRequest('/justifications/direct', { method: 'POST', body: JSON.stringify({ employeeId, question }) })
  notify('انطلب التفسير — يطلعله أول ما يفتح النظام، وجوابه يوصلك بـ«التبريرات» بالصندوق', 'success')
}
