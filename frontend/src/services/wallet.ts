import api from './api'
import type { Wallet, WalletTransaction, CreateTransactionInput } from '../types/wallet'

export const walletService = {
  async get(): Promise<Wallet> {
    const res = await api.get('/wallet')
    return res.data
  },

  async listTransactions(): Promise<WalletTransaction[]> {
    const res = await api.get('/wallet/transactions')
    return res.data
  },

  async createTransaction(data: CreateTransactionInput): Promise<WalletTransaction> {
    const res = await api.post('/wallet/transactions', data)
    return res.data
  },
}
