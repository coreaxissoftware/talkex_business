export interface Wallet {
  id: string
  user_id: string
  balance: number
  currency: string
  created_at: string
  updated_at: string
}

export type TransactionType = 'credit' | 'debit'

export interface WalletTransaction {
  id: string
  wallet_id: string
  type: TransactionType
  amount: number
  balance_after: number
  reference: string | null
  idempotency_key: string
  created_at: string
}

export interface CreateTransactionInput {
  type: TransactionType
  amount: number
  reference?: string | null
  idempotency_key: string
}
