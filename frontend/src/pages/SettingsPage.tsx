import { useState, type FormEvent } from 'react'
import { Settings, User, Lock, Mail, ShieldCheck, Check } from 'lucide-react'
import { useAuthStore } from '../store/authStore'
import { usersService } from '../services/users'
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
    </div>
  )
}
