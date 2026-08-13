import { useEffect, useState } from 'react'
import { List, Plus, Pencil, Trash2, Users, X, Check } from 'lucide-react'
import { contactListsService } from '../services/contactLists'
import { contactsService } from '../services/contacts'
import type { ContactList } from '../types/contactList'
import type { Contact } from '../types/contact'
import Modal from '../components/Modal'

export default function ContactLists() {
  const [lists, setLists] = useState<ContactList[]>([])
  const [contacts, setContacts] = useState<Contact[]>([])
  const [loading, setLoading] = useState(true)

  // Create/Edit modal
  const [showModal, setShowModal] = useState(false)
  const [editingList, setEditingList] = useState<ContactList | null>(null)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [saving, setSaving] = useState(false)

  // Members modal
  const [showMembers, setShowMembers] = useState(false)
  const [selectedList, setSelectedList] = useState<ContactList | null>(null)
  const [memberIds, setMemberIds] = useState<string[]>([])
  const [membersLoading, setMembersLoading] = useState(false)

  // Delete confirm
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const fetchAll = async () => {
    setLoading(true)
    try {
      const [l, c] = await Promise.all([contactListsService.list(), contactsService.list()])
      setLists(l)
      setContacts(c)
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchAll() }, [])

  const openCreate = () => {
    setEditingList(null)
    setName('')
    setDescription('')
    setShowModal(true)
  }

  const openEdit = (l: ContactList) => {
    setEditingList(l)
    setName(l.name)
    setDescription(l.description)
    setShowModal(true)
  }

  const handleSave = async () => {
    if (!name.trim()) return
    setSaving(true)
    try {
      if (editingList) {
        await contactListsService.update(editingList.id, { name, description })
      } else {
        await contactListsService.create({ name, description })
      }
      setShowModal(false)
      await fetchAll()
    } catch {
      // ignore
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await contactListsService.remove(id)
      setDeletingId(null)
      await fetchAll()
    } catch {
      // ignore
    }
  }

  const openMembers = async (l: ContactList) => {
    setSelectedList(l)
    setShowMembers(true)
    setMembersLoading(true)
    try {
      const ids = await contactListsService.getMembers(l.id)
      setMemberIds(ids)
    } catch {
      setMemberIds([])
    } finally {
      setMembersLoading(false)
    }
  }

  const toggleMember = async (contactId: string) => {
    if (!selectedList) return
    if (memberIds.includes(contactId)) {
      await contactListsService.removeMembers(selectedList.id, [contactId])
      setMemberIds(prev => prev.filter(id => id !== contactId))
    } else {
      await contactListsService.addMembers(selectedList.id, [contactId])
      setMemberIds(prev => [...prev, contactId])
    }
    await fetchAll()
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <List size={24} className="text-primary-600" />
            Contact Lists
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Group contacts into lists for targeted campaigns.
          </p>
        </div>
        <button
          onClick={openCreate}
          className="flex items-center gap-1.5 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 transition-colors"
        >
          <Plus size={16} /> New List
        </button>
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-400">Loading...</div>
      ) : lists.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <List size={48} className="mx-auto mb-3 opacity-30" />
          <p className="font-medium">No contact lists yet</p>
          <p className="text-sm mt-1">Create a list to group contacts for campaigns.</p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {lists.map(l => (
            <div
              key={l.id}
              className="rounded-xl border border-gray-200 bg-white p-5 flex flex-col gap-3"
            >
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-semibold text-gray-900">{l.name}</h3>
                  {l.description && (
                    <p className="text-xs text-gray-500 mt-0.5 line-clamp-2">{l.description}</p>
                  )}
                </div>
                <span className="inline-flex items-center gap-1 rounded-full bg-primary-50 px-2.5 py-0.5 text-xs font-semibold text-primary-700">
                  <Users size={12} /> {l.member_count}
                </span>
              </div>

              <div className="flex items-center gap-2 mt-auto pt-2 border-t border-gray-100">
                <button
                  onClick={() => openMembers(l)}
                  className="flex-1 text-xs font-medium text-primary-600 hover:text-primary-700 px-2 py-1.5 rounded-lg hover:bg-primary-50 transition-colors"
                >
                  Manage Members
                </button>
                <button
                  onClick={() => openEdit(l)}
                  className="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
                  title="Edit"
                >
                  <Pencil size={14} />
                </button>
                {deletingId === l.id ? (
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => handleDelete(l.id)}
                      className="p-1.5 rounded-lg text-red-600 hover:bg-red-50 transition-colors"
                      title="Confirm delete"
                    >
                      <Check size={14} />
                    </button>
                    <button
                      onClick={() => setDeletingId(null)}
                      className="p-1.5 rounded-lg text-gray-400 hover:bg-gray-100 transition-colors"
                      title="Cancel"
                    >
                      <X size={14} />
                    </button>
                  </div>
                ) : (
                  <button
                    onClick={() => setDeletingId(l.id)}
                    className="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition-colors"
                    title="Delete"
                  >
                    <Trash2 size={14} />
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create/Edit Modal */}
      {showModal && (
        <Modal title={editingList ? 'Edit List' : 'New Contact List'} onClose={() => setShowModal(false)}>
          <div className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">List Name</label>
              <input
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                placeholder="e.g. VIP Customers"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Description</label>
              <textarea
                value={description}
                onChange={e => setDescription(e.target.value)}
                rows={3}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none resize-none"
                placeholder="Optional description..."
              />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button
                onClick={() => setShowModal(false)}
                className="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleSave}
                disabled={saving || !name.trim()}
                className="px-4 py-2 text-sm font-semibold text-white bg-primary-600 hover:bg-primary-700 rounded-lg disabled:opacity-50 transition-colors"
              >
                {saving ? 'Saving...' : editingList ? 'Update' : 'Create'}
              </button>
            </div>
          </div>
        </Modal>
      )}

      {/* Members Modal */}
      {showMembers && selectedList && (
        <Modal title={`Members — ${selectedList.name}`} onClose={() => setShowMembers(false)}>
          {membersLoading ? (
            <div className="py-8 text-center text-gray-400">Loading...</div>
          ) : contacts.length === 0 ? (
            <div className="py-8 text-center text-gray-400 text-sm">No contacts found. Add contacts first.</div>
          ) : (
            <div className="max-h-80 overflow-y-auto -mx-1">
              {contacts.map(c => {
                const isMember = memberIds.includes(c.id)
                return (
                  <button
                    key={c.id}
                    onClick={() => toggleMember(c.id)}
                    className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-left transition-colors ${
                      isMember
                        ? 'bg-primary-50 hover:bg-primary-100'
                        : 'hover:bg-gray-50'
                    }`}
                  >
                    <div
                      className={`w-5 h-5 rounded border-2 flex items-center justify-center shrink-0 transition-colors ${
                        isMember
                          ? 'bg-primary-600 border-primary-600'
                          : 'border-gray-300'
                      }`}
                    >
                      {isMember && <Check size={12} className="text-white" />}
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-medium text-gray-900 truncate">
                        {c.name || c.phone_number}
                      </p>
                      <p className="text-xs text-gray-500 truncate">{c.phone_number}</p>
                    </div>
                  </button>
                )
              })}
            </div>
          )}
          <div className="flex justify-end pt-4 border-t border-gray-100 mt-4">
            <button
              onClick={() => setShowMembers(false)}
              className="px-4 py-2 text-sm font-semibold text-white bg-primary-600 hover:bg-primary-700 rounded-lg transition-colors"
            >
              Done
            </button>
          </div>
        </Modal>
      )}
    </div>
  )
}
