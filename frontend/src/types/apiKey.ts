export interface ApiKey {
  id: string
  owner_id: string
  name: string
  prefix: string
  last_used_at: string | null
  revoked_at: string | null
  created_at: string
  updated_at: string
}

export interface ApiKeyCreateResult {
  api_key: ApiKey
  plaintext: string
}
