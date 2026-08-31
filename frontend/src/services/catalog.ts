import api from './api'

export interface Product {
  id: string
  owner_id: string
  retailer_id: string
  name: string
  description: string
  image_url: string
  price: number
  currency: string
  availability: string
  url: string
  category: string
  meta_product_id?: string | null
  created_at: string
}

export interface ProductInput {
  retailer_id: string
  name: string
  description?: string
  image_url?: string
  price: number
  currency?: string
  availability?: string
  url?: string
  category?: string
}

export const catalogService = {
  async list(category?: string): Promise<Product[]> {
    const res = await api.get('/catalog', { params: category ? { category } : {} })
    return res.data
  },
  async create(data: ProductInput): Promise<Product> {
    const res = await api.post('/catalog', data)
    return res.data
  },
  async update(id: string, data: Partial<ProductInput>): Promise<Product> {
    const res = await api.put(`/catalog/${id}`, data)
    return res.data
  },
  async remove(id: string): Promise<void> {
    await api.delete(`/catalog/${id}`)
  },
}
