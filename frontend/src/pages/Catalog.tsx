import { useEffect, useState, type FormEvent } from 'react'
import { ShoppingBag, Plus, Trash2, X, IndianRupee, Package } from 'lucide-react'
import { catalogService, type Product, type ProductInput } from '../services/catalog'

const AVAILABILITY = [
  { value: 'in stock', label: 'In stock' },
  { value: 'out of stock', label: 'Out of stock' },
  { value: 'preorder', label: 'Preorder' },
  { value: 'discontinued', label: 'Discontinued' },
]

export default function Catalog() {
  const [items, setItems] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [filter, setFilter] = useState('')

  const [form, setForm] = useState<ProductInput>({
    retailer_id: '',
    name: '',
    description: '',
    image_url: '',
    price: 0,
    currency: 'INR',
    availability: 'in stock',
    url: '',
    category: '',
  })

  const load = async () => {
    setLoading(true)
    try {
      setItems(await catalogService.list())
      setError('')
    } catch {
      setError('Could not load catalog.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    try {
      await catalogService.create(form)
      setForm({
        retailer_id: '',
        name: '',
        description: '',
        image_url: '',
        price: 0,
        currency: 'INR',
        availability: 'in stock',
        url: '',
        category: '',
      })
      setShowForm(false)
      load()
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not create product.')
    }
  }

  const remove = async (p: Product) => {
    if (!confirm(`Delete "${p.name}"?`)) return
    try {
      await catalogService.remove(p.id)
      load()
    } catch {
      setError('Could not delete product.')
    }
  }

  const visible = items.filter(
    (p) =>
      !filter ||
      p.name.toLowerCase().includes(filter.toLowerCase()) ||
      p.retailer_id.toLowerCase().includes(filter.toLowerCase()) ||
      p.category.toLowerCase().includes(filter.toLowerCase())
  )

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <ShoppingBag size={24} className="text-primary-600" /> Product Catalog
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            WhatsApp Commerce catalog · {items.length} product{items.length === 1 ? '' : 's'}
          </p>
        </div>
        <button
          onClick={() => setShowForm(true)}
          className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700"
        >
          <Plus size={16} /> Add product
        </button>
      </div>

      <input
        value={filter}
        onChange={(e) => setFilter(e.target.value)}
        placeholder="Search by name, SKU, or category…"
        className="w-full max-w-md mb-6 rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20"
      />

      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-center py-16 text-gray-500">Loading…</div>
      ) : visible.length === 0 ? (
        <div className="text-center py-16 border-2 border-dashed border-gray-200 rounded-xl">
          <Package size={40} className="mx-auto text-gray-300 mb-3" />
          <p className="text-gray-500">No products yet. Add your first one to enable WhatsApp Commerce.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {visible.map((p) => (
            <div key={p.id} className="rounded-xl border border-gray-200 bg-white p-4 flex flex-col">
              {p.image_url ? (
                <img
                  src={p.image_url}
                  alt={p.name}
                  className="w-full h-40 object-cover rounded-lg mb-3 bg-gray-100"
                  onError={(e) => (e.currentTarget.style.display = 'none')}
                />
              ) : (
                <div className="w-full h-40 rounded-lg mb-3 bg-gray-100 flex items-center justify-center">
                  <Package size={32} className="text-gray-300" />
                </div>
              )}
              <div className="flex-1">
                <div className="flex items-start justify-between gap-2">
                  <h3 className="font-semibold text-gray-900 text-sm">{p.name}</h3>
                  <span
                    className={`text-xs px-2 py-0.5 rounded-full ${
                      p.availability === 'in stock'
                        ? 'bg-green-50 text-green-700'
                        : 'bg-gray-100 text-gray-600'
                    }`}
                  >
                    {p.availability}
                  </span>
                </div>
                <p className="text-xs text-gray-500 mt-1 line-clamp-2">{p.description}</p>
                <p className="text-lg font-bold text-gray-900 mt-2 flex items-center">
                  <IndianRupee size={16} />
                  {p.price.toLocaleString('en-IN')}
                </p>
                <p className="text-xs text-gray-400 mt-1">
                  SKU {p.retailer_id}
                  {p.category && ` · ${p.category}`}
                </p>
              </div>
              <button
                onClick={() => remove(p)}
                className="mt-3 text-xs text-red-600 hover:bg-red-50 rounded px-2 py-1 self-start flex items-center gap-1"
              >
                <Trash2 size={12} /> Delete
              </button>
            </div>
          ))}
        </div>
      )}

      {showForm && (
        <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-xl max-h-[90vh] overflow-y-auto">
            <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between sticky top-0 bg-white">
              <h2 className="font-semibold text-gray-900">Add product</h2>
              <button
                onClick={() => setShowForm(false)}
                className="text-gray-400 hover:text-gray-600"
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={submit} className="p-6 space-y-3">
              <Field label="SKU / Retailer ID" required>
                <input
                  required
                  value={form.retailer_id}
                  onChange={(e) => setForm({ ...form, retailer_id: e.target.value })}
                  className="input"
                  placeholder="SAREE-001"
                />
              </Field>
              <Field label="Name" required>
                <input
                  required
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  className="input"
                />
              </Field>
              <Field label="Description">
                <textarea
                  rows={2}
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  className="input resize-none"
                />
              </Field>
              <div className="grid grid-cols-2 gap-3">
                <Field label="Price (₹)" required>
                  <input
                    required
                    type="number"
                    min={1}
                    step="0.01"
                    value={form.price || ''}
                    onChange={(e) => setForm({ ...form, price: parseFloat(e.target.value) || 0 })}
                    className="input"
                  />
                </Field>
                <Field label="Availability">
                  <select
                    value={form.availability}
                    onChange={(e) => setForm({ ...form, availability: e.target.value })}
                    className="input"
                  >
                    {AVAILABILITY.map((a) => (
                      <option key={a.value} value={a.value}>
                        {a.label}
                      </option>
                    ))}
                  </select>
                </Field>
              </div>
              <Field label="Category">
                <input
                  value={form.category}
                  onChange={(e) => setForm({ ...form, category: e.target.value })}
                  className="input"
                  placeholder="Sarees"
                />
              </Field>
              <Field label="Image URL">
                <input
                  value={form.image_url}
                  onChange={(e) => setForm({ ...form, image_url: e.target.value })}
                  className="input"
                  placeholder="https://…"
                />
              </Field>
              <Field label="Product URL">
                <input
                  value={form.url}
                  onChange={(e) => setForm({ ...form, url: e.target.value })}
                  className="input"
                  placeholder="https://…"
                />
              </Field>
              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowForm(false)}
                  className="px-4 py-2 text-sm rounded-lg text-gray-600 hover:bg-gray-100"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-sm rounded-lg bg-primary-600 text-white hover:bg-primary-700 font-semibold"
                >
                  Save
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      <style>{`
        .input { width: 100%; border-radius: 0.5rem; border: 1px solid #d1d5db;
          padding: 0.5rem 0.75rem; font-size: 0.875rem; outline: none; }
        .input:focus { border-color: #6366f1; box-shadow: 0 0 0 2px rgba(99,102,241,0.2); }
      `}</style>
    </div>
  )
}

function Field({
  label,
  required,
  children,
}: {
  label: string
  required?: boolean
  children: React.ReactNode
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">
        {label} {required && <span className="text-red-500">*</span>}
      </label>
      {children}
    </div>
  )
}
