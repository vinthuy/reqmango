import api from './index'
import type { RecurrenceRule, RecurrenceCreate, RecurrenceUpdate } from '@/types/recurrence'

export async function createRecurrence(issueId: number, data: RecurrenceCreate): Promise<RecurrenceRule> {
  return (await api.post(`/issues/${issueId}/recurrence`, data)).data
}
export async function getRecurrence(issueId: number): Promise<RecurrenceRule> {
  return (await api.get(`/issues/${issueId}/recurrence`)).data
}
export async function updateRecurrence(issueId: number, data: RecurrenceUpdate): Promise<RecurrenceRule> {
  return (await api.put(`/issues/${issueId}/recurrence`, data)).data
}
export async function deleteRecurrence(issueId: number): Promise<void> {
  await api.delete(`/issues/${issueId}/recurrence`)
}
