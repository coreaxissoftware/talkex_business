import { useState, useEffect, type FormEvent } from 'react'
import { Settings, User, Lock, Mail, ShieldCheck, Check, Database, Plus, Trash2, X } from 'lucide-react'
import { useAuthStore } from '../store/authStore'
import { usersService } from '../services/users'
import { customFieldsService } from '../services/customFields'
import type { CustomFieldDefinition } from '../types/customField'
import PasswordInput from '../components/PasswordInput'
import QualityBadge from '../components/QualityBadge'

export default function SettingsPage() {
  const { user, fetchUser } = useAuthStore()

  // Profile form
  const [fullName, setFullName] = useState(user?.full_name ?? '')
  const [businessCategory, setBusinessCategory] = useState(user?.business_category ?? '')
  const [profileSaving, setProfileSaving] = useState(false)
  const [profileError, setProfileError] = useState('')
  const [profileSaved, setProfileSaved] = useState(false)

  // Custom fields
  const [fields, setFields] = useState<CustomFieldDefinition[]>([])
  const [showFieldForm, setShowFieldForm] = useState(false)
  const [fieldName, setFieldName] = useState('')
  const [fieldLabel, setFieldLabel] = useState('')
  const [fieldType, setFieldType] = useState('text')
  const [fieldRequired, setFieldRequired] = useState(false)
  const [fieldOptions, setFieldOptions] = useState('')
  const [fieldSaving, setFieldSaving] = useState(false)
  const [deletingFieldId, setDeletingFieldId] = useState<string | null>(null)

  useEffect(() => {
    customFieldsService.list().then(setFields).catch(() => {})
  }, [])

  const handleFieldCreate = async () => {
    if (!fieldName.trim() || !fieldLabel.trim()) return
    setFieldSaving(true)
    try {
      await customFieldsService.create({
        name: fieldName, label: fieldLabel, type: fieldType,
        required: fieldRequired, options: fieldOptions,
      })
      setShowFieldForm(false)
      setFieldName(''); setFieldLabel(''); setFieldType('text'); setFieldRequired(false); setFieldOptions('')
      setFields(await customFieldsService.list())
    } catch { /* ignore */ }
    finally { setFieldSaving(false) }
  }

  const handleFieldDelete = async (id: string) => {
    try {
      await customFieldsService.remove(id)
      setDeletingFieldId(null)
      setFields(await customFieldsService.list())
    } catch { /* ignore */ }
  }

  // Password form
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [passwordSaving, setPasswordSaving] = useState(false)
  const [passwordError, setPasswordError] = useState('')
  const [passwordSaved, setPasswordSaved] = useState(false)

  const handleProfileSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setProfileError('')
    setProfileSaved(false)
    setProfileSaving(true)
    try {
      await usersService.updateProfile({
        full_name: fullName,
        business_category: businessCategory || null,
      })
      await fetchUser()
      setProfileSaved(true)
      setTimeout(() => setProfileSaved(false), 3000)
    } catch (err: any) {
      setProfileError(err.response?.data?.detail || 'Could not update profile')
    } finally {
      setProfileSaving(false)
    }
  }

  const handlePasswordSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setPasswordError('')
    setPasswordSaved(false)
    if (newPassword !== confirmPassword) {
      setPasswordError('New password and confirmation do not match')
      return
    }
    if (newPassword.length < 8) {
      setPasswordError('New password must be at least 8 characters')
      return
    }
    setPasswordSaving(true)
    try {
      await usersService.changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      })
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
      setPasswordSaved(true)
      setTimeout(() => setPasswordSaved(false), 3000)
    } catch (err: any) {
      setPasswordError(err.response?.data?.detail || 'Could not change password')
    } finally {
      setPasswordSaving(false)
    }
  }

  return (
    <div className="p-6 max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
          <Settings size={24} className="text-primary-600" />
          Settings
        </h1>
        <p className="text-sm text-gray-500 mt-1">Manage your profile and account security.</p>
      </div>

      {/* Profile card */}
      <div className="rounded-xl border border-gray-200 bg-white p-6">
        <h2 className="text-sm font-semibold text-gray-900 flex items-center gap-2 mb-4">
          <User size={16} className="text-primary-600" />
          Profile
        </h2>

        <form onSubmit={handleProfileSubmit} className="space-y-4">
          {profileError && (
            <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
              {profileError}
            </div>
          )}
          {profileSaved && (
            <div className="flex items-center gap-1.5 rounded-lg bg-green-50 border border-green-200 px-3 py-2 text-xs text-green-700">
              <Check size={14} /> Profile updated
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              <span className="inline-flex items-center gap-1">
                <Mail size={12} /> Email
              </span>
            </label>
            <input
              type="email"
              value={user?.email ?? ''}
              disabled
              className="w-full rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-400"
            />
            <p className="text-[11px] text-gray-400 mt-1">Email cannot be changed yet.</p>
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Full Name</label>
            <input
              type="text"
              required
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              Business Category <span className="text-gray-400 font-normal">(optional)</span>
            </label>
            <input
              type="text"
              value={businessCategory}
              onChange={(e) => setBusinessCategory(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="e.g. E-commerce, Healthcare, Education"
            />
          </div>

          <div className="flex items-center gap-2 pt-1">
            {user?.is_business_verified ? (
              <span className="inline-flex items-center gap-1 rounded-full bg-green-50 px-2 py-0.5 text-xs font-semibold text-green-700">
                <ShieldCheck size={12} /> Business Verified
              </span>
            ) : (
              <span className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-semibold text-gray-500">
                Not Verified
              </span>
            )}
            <span className="inline-flex items-center rounded-full bg-primary-50 px-2 py-0.5 text-xs font-semibold text-primary-700 capitalize">
              {user?.role}
            </span>
            {user && <QualityBadge status={user.quality_status} />}
          </div>

          <button
            type="submit"
            disabled={profileSaving}
            className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
          >
            {profileSaving ? 'Saving...' : 'Save Changes'}
          </button>
        </form>
      </div>

      {/* Password card */}
      <div className="rounded-xl border border-gray-200 bg-white p-6">
        <h2 className="text-sm font-semibold text-gray-900 flex items-center gap-2 mb-4">
          <Lock size={16} className="text-primary-600" />
          Change Password
        </h2>

        <form onSubmit={handlePasswordSubmit} className="space-y-4">
          {passwordError && (
            <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
              {passwordError}
            </div>
          )}
          {passwordSaved && (
            <div className="flex items-center gap-1.5 rounded-lg bg-green-50 border border-green-200 px-3 py-2 text-xs text-green-700">
              <Check size={14} /> Password changed
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              Current Password
            </label>
            <PasswordInput value={currentPassword} onChange={setCurrentPassword} minLength={1} />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">New Password</label>
            <PasswordInput
              value={newPassword}
              onChange={setNewPassword}
              placeholder="Minimum 8 characters"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              Confirm New Password
            </label>
            <PasswordInput value={confirmPassword} onChange={setConfirmPassword} />
          </div>

          <button
            type="submit"
            disabled={passwordSaving}
            className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
          >
            {passwordSaving ? 'Updating...' : 'Update Password'}
          </button>
        </form>
      </div>
      {/* Custom Fields card */}
      <div className="rounded-xl border border-gray-200 bg-white p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-sm font-semibold text-gray-900 flex items-center gap-2">
            <Database size={16} className="text-primary-600" />
            Custom Contact Fields
          </h2>
          <button
            onClick={() => setShowFieldForm(true)}
            className="flex items-center gap-1 text-xs font-medium text-primary-600 hover:text-primary-700"
          >
            <Plus size={14} /> Add Field
          </button>
        </div>

        {showFieldForm && (
          <div className="rounded-lg bg-gray-50 border border-gray-200 p-4 mb-4 space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Field Name (slug)</label>
                <input
                  type="text"
                  value={fieldName}
                  onChange={e => setFieldName(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-sm focus:border-primary-500 outline-none"
                  placeholder="e.g. company_name"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Display Label</label>
                <input
                  type="text"
                  value={fieldLabel}
                  onChange={e => setFieldLabel(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-sm focus:border-primary-500 outline-none"
                  placeholder="e.g. Company Name"
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Type</label>
                <select
                  value={fieldType}
                  onChange={e => setFieldType(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-sm focus:border-primary-500 outline-none"
                >
                  <option value="text">Text</option>
                  <option value="number">Number</option>
                  <option value="date">Date</option>
                  <option value="boolean">Boolean (Yes/No)</option>
                  <option value="dropdown">Dropdown</option>
                </select>
              </div>
              {fieldType === 'dropdown' && (
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">Options (comma-separated)</label>
                  <input
                    type="text"
                    value={fieldOptions}
                    onChange={e => setFieldOptions(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-sm focus:border-primary-500 outline-none"
                    placeholder="Option A, Option B, Option C"
                  />
                </div>
              )}
            </div>
            <div className="flex items-center gap-4">
              <label className="flex items-center gap-1.5 text-xs text-gray-700 cursor-pointer">
                <input
                  type="checkbox"
                  checked={fieldRequired}
                  onChange={e => setFieldRequired(e.target.checked)}
                  className="rounded border-gray-300"
                />
                Required
              </label>
              <div className="flex-1" />
              <button
                onClick={() => setShowFieldForm(false)}
                className="px-3 py-1.5 text-xs font-medium text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleFieldCreate}
                disabled={fieldSaving || !fieldName.trim() || !fieldLabel.trim()}
                className="px-3 py-1.5 text-xs font-semibold text-white bg-primary-600 hover:bg-primary-700 rounded-lg disabled:opacity-50 transition-colors"
              >
                {fieldSaving ? 'Saving...' : 'Add'}
              </button>
            </div>
          </div>
        )}

        {fields.length === 0 ? (
          <p className="text-xs text-gray-400 text-center py-4">
            No custom fields defined yet. Add fields to capture extra contact data.
          </p>
        ) : (
          <div className="divide-y divide-gray-100">
            {fields.map(f => (
              <div key={f.id} className="flex items-center justify-between py-2.5">
                <div>
                  <p className="text-sm font-medium text-gray-900">{f.label}</p>
                  <p className="text-xs text-gray-400">
                    <code className="bg-gray-100 px-1 rounded">{f.name}</code>
                    {' '}&middot;{' '}{f.type}
                    {f.required && <span className="text-red-500 ml-1">required</span>}
                    {f.options && <span className="ml-1 text-gray-500">({f.options})</span>}
                  </p>
                </div>
                {deletingFieldId === f.id ? (
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => handleFieldDelete(f.id)}
                      className="p-1 rounded text-red-600 hover:bg-red-50 transition-colors"
                    >
                      <Check size={14} />
                    </button>
                    <button
                      onClick={() => setDeletingFieldId(null)}
                      className="p-1 rounded text-gray-400 hover:bg-gray-100 transition-colors"
                    >
                      <X size={14} />
                    </button>
                  </div>
                ) : (
                  <button
                    onClick={() => setDeletingFieldId(f.id)}
                    className="p-1 rounded text-gray-400 hover:text-red-600 hover:bg-red-50 transition-colors"
                  >
                    <Trash2 size={14} />
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
