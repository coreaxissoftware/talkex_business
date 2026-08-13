export type NotificationType = 'info' | 'success' | 'warning' | 'error'

export interface Notification {
  id: string
  owner_id: string
  type: NotificationType
  title: string
  body: string
  link: string
  read_at: string | null
  created_at: string
  updated_at: string
}
