import api from './api'
import type { AutomationRule, AutomationRuleInput } from '../types/automation'

export const automationService = {
  async list(): Promise<AutomationRule[]> {
    const res = await api.get('/automation/rules')
    return res.data
  },

  async create(data: AutomationRuleInput): Promise<AutomationRule> {
    const res = await api.post('/automation/rules', data)
    return res.data
  },

  async update(id: string, data: Partial<AutomationRuleInput>): Promise<AutomationRule> {
    const res = await api.patch(`/automation/rules/${id}`, data)
    return res.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/automation/rules/${id}`)
  },
}
