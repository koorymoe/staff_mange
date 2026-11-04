import { apiRequest } from './api'

export type WaAgent = 'QUALITY' | 'BOOKING'
export type WaConversation = {
  id: string
  agent: WaAgent
  status: 'BOT' | 'HUMAN' | 'CLOSED'
  customerName: string
  phone: string
  bookingId: string
  bookingCode: string
  sandbox: boolean
  lastMessageAt: string
  lastMessage: string
  unread: number
}
export type WaMessage = {
  id: string
  direction: 'IN' | 'OUT'
  author: 'CUSTOMER' | 'AGENT' | 'HUMAN'
  authorName?: string
  text: string
  createdAt: string
}
export type AiAgentAction = {
  id: string
  agent: WaAgent
  agentName: string
  kind: string
  summary: string
  reason: string
  targetType: 'BOOKING' | 'FOLLOWUP' | 'PROPOSAL'
  targetId: string
  sandbox: boolean
  undoable: boolean
  undoneAt: string | null
  undoneByName: string | null
  createdAt: string
}
export type WaDetail = {
  conversation: WaConversation
  messages: WaMessage[]
  actions: AiAgentAction[]
}
export const allowedWaAgents = (
  role: string | undefined,
  actualRole: string | undefined,
  permissions: string[],
): WaAgent[] =>
  actualRole === 'OWNER' || role === 'ADMIN'
    ? ['QUALITY', 'BOOKING']
    : [
        ...(permissions.includes('quality_control')
          ? ['QUALITY' as const]
          : []),
        ...(permissions.includes('coordinator') ? ['BOOKING' as const] : []),
      ]
const path = (id: string) => `/wa/conversations/${encodeURIComponent(id)}`
export const waApi = {
  list: (agent: WaAgent, status: string) =>
    apiRequest<WaConversation[]>(
      `/wa/conversations?agent=${agent}${status ? `&status=${encodeURIComponent(status)}` : ''}`,
    ),
  detail: (id: string) => apiRequest<WaDetail>(path(id)),
  takeover: (id: string) =>
    apiRequest(`${path(id)}/takeover`, { method: 'POST' }),
  release: (id: string) =>
    apiRequest(`${path(id)}/release`, { method: 'POST' }),
  close: (id: string) => apiRequest(`${path(id)}/close`, { method: 'POST' }),
  send: (id: string, text: string) =>
    apiRequest<WaMessage>(`${path(id)}/send`, {
      method: 'POST',
      body: JSON.stringify({ text }),
    }),
  actions: (targetId: string) =>
    apiRequest<AiAgentAction[]>(
      `/wa/actions?targetId=${encodeURIComponent(targetId)}`,
    ),
  undo: (id: string, reason: string) =>
    apiRequest<AiAgentAction>(`/wa/actions/${encodeURIComponent(id)}/undo`, {
      method: 'POST',
      body: JSON.stringify({ reason }),
    }),
  sandboxStart: (agent: WaAgent, bookingId: string) =>
    apiRequest<{ conversation: WaConversation; messages: WaMessage[] }>(
      '/wa/sandbox/start',
      { method: 'POST', body: JSON.stringify({ agent, bookingId }) },
    ),
  sandboxSay: (id: string, text: string) =>
    apiRequest<{ messages: WaMessage[]; actions: AiAgentAction[] }>(
      `/wa/sandbox/${encodeURIComponent(id)}/say`,
      { method: 'POST', body: JSON.stringify({ text }) },
    ),
}
