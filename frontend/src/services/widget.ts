import api from './api'

export interface WidgetConfig {
  id: string
  owner_id: string
  public_key: string
  enabled: boolean
  title: string
  greeting: string
  theme_color: string
  created_at: string
  updated_at: string
}

export const widgetService = {
  async get(): Promise<WidgetConfig> {
    const res = await api.get('/settings/widget')
    return res.data
  },
  async update(data: Partial<Pick<WidgetConfig, 'enabled' | 'title' | 'greeting' | 'theme_color'>>): Promise<WidgetConfig> {
    const res = await api.patch('/settings/widget', data)
    return res.data
  },
  async rotateKey(): Promise<WidgetConfig> {
    const res = await api.post('/settings/widget/rotate-key')
    return res.data
  },
}
