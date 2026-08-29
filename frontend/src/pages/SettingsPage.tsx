import { useState, useEffect, type FormEvent } from 'react'
import { Settings, User, Lock, Mail, ShieldCheck, Check, Database, Plus, Trash2, X, AlertTriangle, Bell, Building2, Smartphone, Monitor, LogOut } from 'lucide-react'
import { useAuthStore } from '../store/authStore'
import { usersService, type Session } from '../services/users'
import { customFieldsService } from '../services/customFields'
import { customersService } from '../services/customers'
import { settingsService, type UserPrefs } from '../services/settings'
import type { CustomFieldDefinition } from '../types/customField'
import type { Customer, CustomerUpsertInput } from '../types/customer'
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

  // Business profile
  const [biz, setBiz] = useState<Customer | null>(null)
  const [bizForm, setBizForm] = useState<CustomerUpsertInput>({ business_name: '', business_category: '' })
  const [bizSaving, setBizSaving] = useState(false)
  const [bizSaved, setBizSaved] = useState(false)
  const [bizError, setBizError] = useState('')

  const loadSessions = () => usersService.listSessions().then(setSessions).catch(() => {})

  useEffect(() => {
    customFieldsService.list().then(setFields).catch(() => {})
    settingsService.get().then(setNotifPrefs).catch(() => {})
    loadSessions()
    customersService.get().then((c) => {
      setBiz(c)
      setBizForm({
        business_name: c.business_name,
        business_category: c.business_category,
        gstin: c.gstin,
        website: c.website,
        address: c.address,
        city: c.city,
        state: c.state,
        country: c.country,
        phone: c.phone,
      })
    }).catch(() => {})
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

  // Notification preferences (backend-persisted)
  const [notifPrefs, setNotifPrefs] = useState<UserPrefs>({
    notif_campaigns: true,
    notif_messages: true,
    notif_system: true,
    email_digest: false,
    timezone: 'Asia/Kolkata',
    language: 'en',
    auto_pause_enabled: false,
    min_balance: 0,
    sandbox_mode: false,
    approval_threshold: 0,
    cost_whatsapp: 0,
    cost_sms: 0,
    cost_talkex: 0,
    cost_telegram: 0,
    cost_email: 0,
    cost_rcs: 0,
    cost_instagram: 0,
    cost_messenger: 0,
    sell_whatsapp: 0,
    sell_sms: 0,
    sell_talkex: 0,
    sell_telegram: 0,
    sell_email: 0,
    sell_rcs: 0,
    sell_instagram: 0,
    sell_messenger: 0,
    business_hours_enabled: false,
    business_days: [1, 2, 3, 4, 5],
    business_open_time: '09:00',
    business_close_time: '18:00',
    away_message: 'Thanks for your message! We are currently offline. We\'ll get back to you during business hours.',
    sla_first_response_mins: 0,
    ai_auto_tag_enabled: false,
  })
  const [notifSaving, setNotifSaving] = useState(false)
  const [notifSaved, setNotifSaved] = useState(false)

  // 2FA
  const [tfaSecret, setTfaSecret] = useState('')
  const [tfaUri, setTfaUri] = useState('')
  const [tfaCode, setTfaCode] = useState('')
  const [tfaLoading, setTfaLoading] = useState(false)
  const [tfaError, setTfaError] = useState('')
  const [tfaDisablePassword, setTfaDisablePassword] = useState('')
  const [tfaDisableCode, setTfaDisableCode] = useState('')

  // Sessions
  const [sessions, setSessions] = useState<Session[]>([])
  const [sessionsLoading, setSessionsLoading] = useState(false)

  // Danger zone
  const [deactivatePassword, setDeactivatePassword] = useState('')
  const [deactivating, setDeactivating] = useState(false)
  const [deactivateError, setDeactivateError] = useState('')
  const [showDeactivate, setShowDeactivate] = useState(false)

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

      {/* Business Profile card */}
      <div className="rounded-xl border border-gray-200 bg-white p-6">
        <h2 className="text-sm font-semibold text-gray-900 flex items-center gap-2 mb-4">
          <Building2 size={16} className="text-primary-600" />
          Business Profile
          {biz && (
            <span className={`ml-auto inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-semibold ${
              biz.verification_status === 'verified' ? 'bg-green-50 text-green-700' :
              biz.verification_status === 'rejected' ? 'bg-red-50 text-red-700' :
              'bg-yellow-50 text-yellow-700'
            }`}>
              {biz.verification_status}
            </span>
          )}
        </h2>

        <form onSubmit={async (e: FormEvent) => {
          e.preventDefault()
          setBizError(''); setBizSaved(false); setBizSaving(true)
          try {
            const result = await customersService.upsert(bizForm)
            setBiz(result); setBizSaved(true)
            setTimeout(() => setBizSaved(false), 3000)
          } catch (err: any) {
            setBizError(err.response?.data?.detail || 'Could not save business profile')
          } finally { setBizSaving(false) }
        }} className="space-y-4">
          {bizError && <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">{bizError}</div>}
          {bizSaved && <div className="flex items-center gap-1.5 rounded-lg bg-green-50 border border-green-200 px-3 py-2 text-xs text-green-700"><Check size={14} /> Business profile saved</div>}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Business Name *</label>
              <input type="text" required value={bizForm.business_name}
                onChange={e => setBizForm(p => ({ ...p, business_name: e.target.value }))}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                placeholder="CoreAxis Ventures" />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Category</label>
              <input type="text" value={bizForm.business_category}
                onChange={e => setBizForm(p => ({ ...p, business_category: e.target.value }))}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                placeholder="Technology, E-commerce" />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">GSTIN</label>
              <input type="text" value={bizForm.gstin ?? ''}
                onChange={e => setBizForm(p => ({ ...p, gstin: e.target.value || null }))}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                placeholder="22AAAAA0000A1Z5" />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Phone</label>
              <input type="tel" value={bizForm.phone ?? ''}
                onChange={e => setBizForm(p => ({ ...p, phone: e.target.value || null }))}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                placeholder="+91 98765 43210" />
            </div>
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Website</label>
            <input type="url" value={bizForm.website ?? ''}
              onChange={e => setBizForm(p => ({ ...p, website: e.target.value || null }))}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="https://coreaxis.cloud" />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Address</label>
            <input type="text" value={bizForm.address ?? ''}
              onChange={e => setBizForm(p => ({ ...p, address: e.target.value || null }))}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="123 Business Park, Sector 5" />
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">City</label>
              <input type="text" value={bizForm.city ?? ''}
                onChange={e => setBizForm(p => ({ ...p, city: e.target.value || null }))}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none" />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">State</label>
              <input type="text" value={bizForm.state ?? ''}
                onChange={e => setBizForm(p => ({ ...p, state: e.target.value || null }))}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none" />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Country</label>
              <input type="text" value={bizForm.country ?? 'IN'}
                onChange={e => setBizForm(p => ({ ...p, country: e.target.value }))}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none" />
            </div>
          </div>

          <button type="submit" disabled={bizSaving}
            className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors">
            {bizSaving ? 'Saving...' : 'Save Business Profile'}
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
      {/* Two-Factor Authentication */}
      <div className="rounded-xl border border-gray-200 bg-white p-6">
        <h2 className="text-sm font-semibold text-gray-900 flex items-center gap-2 mb-4">
          <Smartphone size={16} className="text-primary-600" />
          Two-Factor Authentication
          {user?.two_factor_enabled ? (
            <span className="ml-auto inline-flex items-center rounded-full bg-green-50 px-2 py-0.5 text-[10px] font-semibold text-green-700">Enabled</span>
          ) : (
            <span className="ml-auto inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-[10px] font-semibold text-gray-500">Disabled</span>
          )}
        </h2>

        {tfaError && <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700 mb-3">{tfaError}</div>}

        {!user?.two_factor_enabled ? (
          <>
            {!tfaSecret ? (
              <div className="space-y-3">
                <p className="text-sm text-gray-600">Add an extra layer of security by enabling TOTP-based two-factor authentication.</p>
                <button
                  onClick={async () => {
                    setTfaLoading(true); setTfaError('')
                    try {
                      const res = await usersService.setup2FA()
                      setTfaSecret(res.secret)
                      setTfaUri(res.provisioning_uri)
                    } catch (err: any) {
                      setTfaError(err.response?.data?.detail || 'Failed to set up 2FA')
                    } finally { setTfaLoading(false) }
                  }}
                  disabled={tfaLoading}
                  className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
                >
                  {tfaLoading ? 'Setting up...' : 'Enable 2FA'}
                </button>
              </div>
            ) : (
              <div className="space-y-4">
                <p className="text-sm text-gray-600">Scan this secret in your authenticator app (Google Authenticator, Authy, etc.):</p>
                <div className="rounded-lg bg-gray-50 border border-gray-200 p-3">
                  <p className="text-xs text-gray-500 mb-1">Secret key (manual entry):</p>
                  <code className="text-sm font-mono text-gray-900 break-all select-all">{tfaSecret}</code>
                </div>
                <div className="rounded-lg bg-gray-50 border border-gray-200 p-3">
                  <p className="text-xs text-gray-500 mb-1">Provisioning URI:</p>
                  <code className="text-[11px] font-mono text-gray-700 break-all select-all">{tfaUri}</code>
                </div>
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1.5">Enter code from your app to verify</label>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={tfaCode}
                      onChange={e => setTfaCode(e.target.value)}
                      maxLength={6}
                      placeholder="000000"
                      className="w-32 rounded-lg border border-gray-300 px-3 py-2 text-sm font-mono text-center tracking-widest focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                    />
                    <button
                      onClick={async () => {
                        setTfaLoading(true); setTfaError('')
                        try {
                          await usersService.verify2FA(tfaCode)
                          await fetchUser()
                          setTfaSecret(''); setTfaUri(''); setTfaCode('')
                        } catch (err: any) {
                          setTfaError(err.response?.data?.detail || 'Invalid code')
                        } finally { setTfaLoading(false) }
                      }}
                      disabled={tfaLoading || tfaCode.length < 6}
                      className="rounded-lg bg-green-600 px-4 py-2 text-sm font-semibold text-white hover:bg-green-700 disabled:opacity-50 transition-colors"
                    >
                      {tfaLoading ? 'Verifying...' : 'Verify & Enable'}
                    </button>
                  </div>
                </div>
              </div>
            )}
          </>
        ) : (
          <div className="space-y-4">
            <p className="text-sm text-gray-600">To disable 2FA, enter your password and a current code from your authenticator app.</p>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1.5">Password</label>
                <PasswordInput value={tfaDisablePassword} onChange={setTfaDisablePassword} />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1.5">2FA Code</label>
                <input
                  type="text"
                  value={tfaDisableCode}
                  onChange={e => setTfaDisableCode(e.target.value)}
                  maxLength={6}
                  placeholder="000000"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm font-mono tracking-widest focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                />
              </div>
            </div>
            <button
              onClick={async () => {
                setTfaLoading(true); setTfaError('')
                try {
                  await usersService.disable2FA(tfaDisablePassword, tfaDisableCode)
                  await fetchUser()
                  setTfaDisablePassword(''); setTfaDisableCode('')
                } catch (err: any) {
                  setTfaError(err.response?.data?.detail || 'Failed to disable 2FA')
                } finally { setTfaLoading(false) }
              }}
              disabled={tfaLoading || !tfaDisablePassword || tfaDisableCode.length < 6}
              className="rounded-lg border border-red-300 px-4 py-2 text-sm font-medium text-red-700 hover:bg-red-50 disabled:opacity-50 transition-colors"
            >
              {tfaLoading ? 'Disabling...' : 'Disable 2FA'}
            </button>
          </div>
        )}
      </div>

      {/* Active Sessions */}
      <div className="rounded-xl border border-gray-200 bg-white p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-sm font-semibold text-gray-900 flex items-center gap-2">
            <Monitor size={16} className="text-primary-600" />
            Active Sessions
          </h2>
          {sessions.length > 1 && (
            <button
              onClick={async () => {
                setSessionsLoading(true)
                try {
                  await usersService.revokeAllSessions()
                  await loadSessions()
                } catch { /* ignore */ }
                finally { setSessionsLoading(false) }
              }}
              disabled={sessionsLoading}
              className="flex items-center gap-1 text-xs font-medium text-red-600 hover:text-red-700"
            >
              <LogOut size={12} /> Revoke All
            </button>
          )}
        </div>

        {sessions.length === 0 ? (
          <p className="text-xs text-gray-400 text-center py-4">No active sessions found.</p>
        ) : (
          <div className="divide-y divide-gray-100">
            {sessions.map(s => (
              <div key={s.id} className="flex items-center justify-between py-2.5">
                <div>
                  <p className="text-sm font-medium text-gray-900 truncate max-w-xs" title={s.user_agent}>
                    {s.user_agent ? s.user_agent.substring(0, 60) + (s.user_agent.length > 60 ? '...' : '') : 'Unknown device'}
                  </p>
                  <p className="text-xs text-gray-400">
                    {s.ip_address || 'Unknown IP'} &middot; Since {new Date(s.created_at).toLocaleDateString()}
                  </p>
                </div>
                <button
                  onClick={async () => {
                    try {
                      await usersService.revokeSession(s.id)
                      await loadSessions()
                    } catch { /* ignore */ }
                  }}
                  className="p-1.5 rounded text-gray-400 hover:text-red-600 hover:bg-red-50 transition-colors"
                  title="Revoke session"
                >
                  <X size={14} />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Notification Preferences */}
      <div className="rounded-xl border border-gray-200 bg-white p-6">
        <h2 className="text-sm font-semibold text-gray-900 flex items-center gap-2 mb-4">
          <Bell size={16} className="text-primary-600" />
          Notification Preferences
        </h2>
        {notifSaved && <div className="flex items-center gap-1.5 rounded-lg bg-green-50 border border-green-200 px-3 py-2 text-xs text-green-700 mb-3"><Check size={14} /> Preferences saved</div>}
        <div className="space-y-3">
          {[
            { key: 'notif_campaigns' as const, label: 'Campaign Completions', desc: 'When a campaign finishes sending' },
            { key: 'notif_messages' as const, label: 'Inbound Messages', desc: 'When a contact sends you a message' },
            { key: 'notif_system' as const, label: 'System Alerts', desc: 'Low balance, quality warnings, errors' },
            { key: 'email_digest' as const, label: 'Email Digest', desc: 'Daily summary email of platform activity' },
          ].map(item => (
            <label key={item.key} className="flex items-center justify-between py-2 cursor-pointer">
              <div>
                <p className="text-sm font-medium text-gray-900">{item.label}</p>
                <p className="text-xs text-gray-500">{item.desc}</p>
              </div>
              <input
                type="checkbox"
                checked={notifPrefs[item.key]}
                onChange={e => setNotifPrefs(prev => ({ ...prev, [item.key]: e.target.checked }))}
                className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              />
            </label>
          ))}
        </div>
        <div className="border-t border-gray-100 mt-4 pt-4">
          <h3 className="text-xs font-semibold text-gray-700 mb-3">Wallet & Campaign Safety</h3>
          <label className="flex items-center justify-between py-2 cursor-pointer">
            <div>
              <p className="text-sm font-medium text-gray-900">Auto-Pause on Low Balance</p>
              <p className="text-xs text-gray-500">Automatically pause running campaigns when wallet balance drops below threshold</p>
            </div>
            <input
              type="checkbox"
              checked={notifPrefs.auto_pause_enabled}
              onChange={e => setNotifPrefs(prev => ({ ...prev, auto_pause_enabled: e.target.checked }))}
              className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
          </label>
          <label className="flex items-center justify-between py-2 cursor-pointer">
            <div>
              <p className="text-sm font-medium text-gray-900">Sandbox Mode (Developer Testing)</p>
              <p className="text-xs text-gray-500">Messages are simulated, not actually delivered. For testing only.</p>
            </div>
            <input
              type="checkbox"
              checked={notifPrefs.sandbox_mode}
              onChange={e => setNotifPrefs(prev => ({ ...prev, sandbox_mode: e.target.checked }))}
              className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
          </label>
          {notifPrefs.sandbox_mode && (
            <div className="rounded-lg bg-amber-50 border border-amber-200 px-3 py-2 text-xs text-amber-700 flex items-center gap-2">
              ⚠️ Sandbox mode is active — all messages will be simulated, not delivered to real recipients.
            </div>
          )}
          {notifPrefs.auto_pause_enabled && (
            <div className="ml-0 mt-2">
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Minimum Balance (INR)</label>
              <input
                type="number"
                min="0"
                step="10"
                value={notifPrefs.min_balance}
                onChange={e => setNotifPrefs(prev => ({ ...prev, min_balance: parseFloat(e.target.value) || 0 }))}
                className="w-40 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                placeholder="e.g. 100"
              />
              <p className="text-[11px] text-gray-400 mt-1">Campaigns auto-resume when balance is recharged above this amount.</p>
            </div>
          )}
        </div>
        <div className="border-t border-gray-100 mt-4 pt-4">
          <h3 className="text-xs font-semibold text-gray-700 mb-3">Campaign Approval</h3>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Approval Threshold (recipients)</label>
            <input
              type="number"
              min="0"
              step="1"
              value={notifPrefs.approval_threshold}
              onChange={e => setNotifPrefs(prev => ({ ...prev, approval_threshold: parseInt(e.target.value) || 0 }))}
              className="w-40 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="e.g. 1000"
            />
            <p className="text-[11px] text-gray-400 mt-1">Campaigns with recipients above this number require approval before launch. Set to 0 to disable.</p>
          </div>
        </div>

        <div className="border-t border-gray-100 mt-4 pt-4">
          <h3 className="text-xs font-semibold text-gray-700 mb-3">Message Pricing (per message, INR)</h3>
          <div className="grid grid-cols-4 gap-4">
            {[
              { label: 'WhatsApp', costKey: 'cost_whatsapp' as const, sellKey: 'sell_whatsapp' as const },
              { label: 'SMS', costKey: 'cost_sms' as const, sellKey: 'sell_sms' as const },
              { label: 'TalkEx', costKey: 'cost_talkex' as const, sellKey: 'sell_talkex' as const },
              { label: 'Telegram', costKey: 'cost_telegram' as const, sellKey: 'sell_telegram' as const },
              { label: 'Email', costKey: 'cost_email' as const, sellKey: 'sell_email' as const },
              { label: 'RCS', costKey: 'cost_rcs' as const, sellKey: 'sell_rcs' as const },
              { label: 'Instagram', costKey: 'cost_instagram' as const, sellKey: 'sell_instagram' as const },
              { label: 'Messenger', costKey: 'cost_messenger' as const, sellKey: 'sell_messenger' as const },
            ].map(ch => (
              <div key={ch.costKey} className="space-y-2">
                <p className="text-xs font-medium text-gray-700">{ch.label}</p>
                <div>
                  <label className="block text-[11px] text-gray-500 mb-0.5">Cost (your cost)</label>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    value={notifPrefs[ch.costKey]}
                    onChange={e => setNotifPrefs(prev => ({ ...prev, [ch.costKey]: parseFloat(e.target.value) || 0 }))}
                    className="w-full rounded-lg border border-gray-300 px-2 py-1.5 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                  />
                </div>
                <div>
                  <label className="block text-[11px] text-gray-500 mb-0.5">Sell (charge client)</label>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    value={notifPrefs[ch.sellKey]}
                    onChange={e => setNotifPrefs(prev => ({ ...prev, [ch.sellKey]: parseFloat(e.target.value) || 0 }))}
                    className="w-full rounded-lg border border-gray-300 px-2 py-1.5 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                  />
                </div>
              </div>
            ))}
          </div>
          <p className="text-[11px] text-gray-400 mt-2">Set per-message cost and sell price for margin tracking. Leave at 0 to disable.</p>
        </div>

        {/* Business Hours */}
        <div className="mt-6 pt-6 border-t border-gray-100">
          <h3 className="text-sm font-semibold text-gray-900 mb-3">Business Hours & Auto-Reply</h3>
          <label className="flex items-center justify-between py-2 cursor-pointer">
            <div>
              <p className="text-sm font-medium text-gray-900">Enable business hours</p>
              <p className="text-xs text-gray-500">Auto-reply with your away message when a customer messages outside hours.</p>
            </div>
            <input
              type="checkbox"
              checked={notifPrefs.business_hours_enabled}
              onChange={e => setNotifPrefs(prev => ({ ...prev, business_hours_enabled: e.target.checked }))}
              className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
          </label>
          {notifPrefs.business_hours_enabled && (
            <div className="ml-0 mt-3 space-y-3">
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1.5">Working days</label>
                <div className="flex gap-1 flex-wrap">
                  {[
                    { d: 1, label: 'Mon' }, { d: 2, label: 'Tue' }, { d: 3, label: 'Wed' },
                    { d: 4, label: 'Thu' }, { d: 5, label: 'Fri' }, { d: 6, label: 'Sat' },
                    { d: 7, label: 'Sun' },
                  ].map(({ d, label }) => {
                    const on = notifPrefs.business_days.includes(d)
                    return (
                      <button
                        key={d}
                        type="button"
                        onClick={() => setNotifPrefs(prev => ({
                          ...prev,
                          business_days: on
                            ? prev.business_days.filter(x => x !== d)
                            : [...prev.business_days, d].sort(),
                        }))}
                        className={`rounded-lg border px-3 py-1 text-xs font-medium ${
                          on
                            ? 'border-primary-500 bg-primary-50 text-primary-700'
                            : 'border-gray-200 text-gray-500 hover:border-gray-300'
                        }`}
                      >
                        {label}
                      </button>
                    )
                  })}
                </div>
              </div>
              <div className="flex gap-3">
                <div className="flex-1">
                  <label className="block text-xs font-medium text-gray-700 mb-1.5">Open</label>
                  <input
                    type="time"
                    value={notifPrefs.business_open_time}
                    onChange={e => setNotifPrefs(prev => ({ ...prev, business_open_time: e.target.value }))}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500"
                  />
                </div>
                <div className="flex-1">
                  <label className="block text-xs font-medium text-gray-700 mb-1.5">Close</label>
                  <input
                    type="time"
                    value={notifPrefs.business_close_time}
                    onChange={e => setNotifPrefs(prev => ({ ...prev, business_close_time: e.target.value }))}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500"
                  />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1.5">Away message</label>
                <textarea
                  rows={3}
                  value={notifPrefs.away_message}
                  onChange={e => setNotifPrefs(prev => ({ ...prev, away_message: e.target.value }))}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500"
                />
                <p className="text-[11px] text-gray-400 mt-1">Sent once per contact per 24h so a chatty visitor isn't spammed.</p>
              </div>
            </div>
          )}
        </div>

        {/* SLA */}
        <div className="mt-6 pt-6 border-t border-gray-100">
          <h3 className="text-sm font-semibold text-gray-900 mb-3">SLA policy</h3>
          <label className="block text-xs font-medium text-gray-700 mb-1.5">First-response threshold (minutes)</label>
          <div className="flex items-center gap-2">
            <input
              type="number"
              min="0"
              value={notifPrefs.sla_first_response_mins}
              onChange={e => setNotifPrefs(prev => ({ ...prev, sla_first_response_mins: parseInt(e.target.value) || 0 }))}
              className="w-32 rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500"
            />
            <span className="text-xs text-gray-500">0 = disabled. Alerts fire when a conversation goes unanswered past this window.</span>
          </div>
        </div>

        {/* AI Auto-tag */}
        <div className="mt-6 pt-6 border-t border-gray-100">
          <h3 className="text-sm font-semibold text-gray-900 mb-3">AI features</h3>
          <label className="flex items-center justify-between py-2 cursor-pointer">
            <div>
              <p className="text-sm font-medium text-gray-900">Auto-tag inbound with sentiment</p>
              <p className="text-xs text-gray-500">Runs Claude on every inbound and stamps `sentiment:positive/neutral/negative` on the conversation labels. Needs ANTHROPIC_API_KEY for real classification; dev heuristic otherwise.</p>
            </div>
            <input
              type="checkbox"
              checked={notifPrefs.ai_auto_tag_enabled}
              onChange={e => setNotifPrefs(prev => ({ ...prev, ai_auto_tag_enabled: e.target.checked }))}
              className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
          </label>
        </div>

        <button
          onClick={async () => {
            setNotifSaving(true); setNotifSaved(false)
            try { await settingsService.save(notifPrefs); setNotifSaved(true); setTimeout(() => setNotifSaved(false), 3000) }
            catch { /* ignore */ }
            finally { setNotifSaving(false) }
          }}
          disabled={notifSaving}
          className="mt-4 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
        >
          {notifSaving ? 'Saving...' : 'Save Preferences'}
        </button>
      </div>

      {/* Danger Zone */}
      <div className="rounded-xl border border-red-200 bg-white p-6">
        <h2 className="text-sm font-semibold text-red-700 flex items-center gap-2 mb-4">
          <AlertTriangle size={16} />
          Danger Zone
        </h2>

        {!showDeactivate ? (
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-900">Deactivate Account</p>
              <p className="text-xs text-gray-500">Once deactivated, you will be logged out and unable to sign in.</p>
            </div>
            <button
              onClick={() => setShowDeactivate(true)}
              className="rounded-lg border border-red-300 px-4 py-2 text-sm font-medium text-red-700 hover:bg-red-50 transition-colors"
            >
              Deactivate
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {deactivateError && (
              <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">{deactivateError}</div>
            )}
            <p className="text-sm text-gray-700">Enter your password to confirm account deactivation. This cannot be undone.</p>
            <PasswordInput value={deactivatePassword} onChange={setDeactivatePassword} placeholder="Current password" />
            <div className="flex gap-2">
              <button
                onClick={() => { setShowDeactivate(false); setDeactivatePassword(''); setDeactivateError('') }}
                className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                disabled={deactivating || !deactivatePassword}
                onClick={async () => {
                  setDeactivating(true)
                  setDeactivateError('')
                  try {
                    await usersService.deactivateAccount(deactivatePassword)
                    const { logout } = useAuthStore.getState()
                    logout()
                  } catch (err: any) {
                    setDeactivateError(err.response?.data?.detail || 'Could not deactivate account')
                  } finally {
                    setDeactivating(false)
                  }
                }}
                className="flex-1 rounded-lg bg-red-600 px-4 py-2 text-sm font-semibold text-white hover:bg-red-700 disabled:opacity-50 transition-colors"
              >
                {deactivating ? 'Deactivating...' : 'Confirm Deactivate'}
              </button>
            </div>
          </div>
        )}
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
