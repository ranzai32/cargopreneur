import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button } from '../components/Button'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageHeader } from '../components/PageHeader'
import { PageLayout } from '../components/PageLayout'
import { TextField } from '../components/TextField'
import type { ProfileInput } from '../lib/api'
import { fieldErrorsOf, isApiError, useAppState } from '../lib/app-state'
import { formatDeliveryStatus, formatNumber, formatQuantity } from '../lib/format'
import type { ApiOrder, ApiUser } from '../types/catalog'
import { StatusBadge } from './OrdersPage'

export function AccountPage() {
  const { user, ready, logout, loadOrders } = useAppState()
  const navigate = useNavigate()
  const [orders, setOrders] = useState<ApiOrder[]>([])
  const [editing, setEditing] = useState(false)

  useEffect(() => {
    document.title = 'My account · CargoPrener'
  }, [])

  useEffect(() => {
    if (!user) return
    loadOrders('all')
      .then((items) => setOrders(items.filter((item) => item.yourQuantity)))
      .catch(() => setOrders([]))
  }, [loadOrders, user])

  async function handleLogout() {
    await logout()
    navigate('/')
  }

  const active = orders.filter((item) => item.status === 'open' || item.status === 'confirmed').length
  const awaitingPayment = orders.filter((item) => item.status === 'confirmed' && item.yourPaymentStatus !== 'paid').length

  return (
    <PageLayout>
      <Container className="py-section">
        <PageHeader title="My account" lede="Your business profile and the group orders you are part of." />
        {!ready && <p className="mt-8 text-ink-700">Loading…</p>}
        {ready && !user && (
          <Card className="mt-8 max-w-xl p-6">
            <p className="text-ink-800">Sign in to see your account.</p>
            <div className="mt-4 flex gap-3">
              <Button to="/login">Log in</Button>
              <Button to="/register" variant="secondary">Create an account</Button>
            </div>
          </Card>
        )}
        {ready && user && (
          <div className="mt-8 grid gap-6 lg:grid-cols-[2fr_3fr]">
            <Card className="p-6">
              {editing ? (
                <ProfileForm user={user} onDone={() => setEditing(false)} />
              ) : (
                <>
                  <div className="flex items-center gap-4">
                    <span aria-hidden="true" className="inline-flex size-14 items-center justify-center rounded-full bg-accent-100 font-serif text-xl text-accent-800">
                      {initials(user.businessName)}
                    </span>
                    <div className="min-w-0">
                      <h2 className="truncate font-serif text-2xl text-ink-950">{user.businessName}</h2>
                      <p className="truncate text-sm text-ink-700">{user.email}</p>
                    </div>
                  </div>
                  <dl className="mt-6 grid gap-3 text-sm">
                    <ProfileRow label="Contact name" value={user.contactName} />
                    <ProfileRow label="Region" value={user.region} />
                    <ProfileRow label="Member since" value={new Date(user.createdAt).toLocaleDateString('en-US', { dateStyle: 'medium' })} />
                  </dl>
                  <div className="mt-6 flex flex-wrap gap-3">
                    <Button onClick={() => setEditing(true)}>Edit profile</Button>
                    <Button variant="secondary" onClick={() => void handleLogout()}>Log out</Button>
                  </div>
                </>
              )}
            </Card>
            <Card id="orders" className="p-6">
              <h2 className="font-serif text-2xl text-ink-950">Your orders</h2>
              <dl className="mt-4 grid grid-cols-3 gap-3">
                <Stat label="Joined" value={orders.length} />
                <Stat label="Active" value={active} />
                <Stat label="To pay" value={awaitingPayment} />
              </dl>
              {orders.length === 0 ? (
                <p className="mt-6 text-sm text-ink-700">
                  You have not joined any orders yet. <Link to="/orders" className="font-semibold text-accent-800">Browse open orders</Link>
                </p>
              ) : (
                <ul className="mt-6 divide-y divide-ink-200">
                  {orders.map((order) => (
                    <li key={order.id}>
                      <Link to={`/orders/${order.id}`} className="flex items-center justify-between gap-3 py-3 hover:text-accent-800">
                        <span className="min-w-0">
                          <span className="block truncate font-semibold text-ink-900">{order.product.name}</span>
                          <span className="text-sm text-ink-700">
                            {formatQuantity(order.yourQuantity ?? 0, order.product.unit)}
                            {' · '}
                            {order.yourPaymentStatus === 'paid' ? `Paid · ${formatDeliveryStatus(order.yourDeliveryStatus)}` : 'Unpaid'}
                          </span>
                        </span>
                        <StatusBadge status={order.status} />
                      </Link>
                    </li>
                  ))}
                </ul>
              )}
            </Card>
          </div>
        )}
      </Container>
    </PageLayout>
  )
}

function ProfileForm({ user, onDone }: { user: ApiUser; onDone: () => void }) {
  const { updateProfile } = useAppState()
  const [form, setForm] = useState<ProfileInput>({
    businessName: user.businessName,
    contactName: user.contactName,
    email: user.email,
    region: user.region,
  })
  const [message, setMessage] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)

  function update(key: keyof ProfileInput, value: string) {
    setForm((current) => ({ ...current, [key]: value }))
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const nextErrors: Record<string, string> = {}
    if (!form.businessName.trim()) nextErrors.businessName = 'Business name is required'
    if (!form.contactName.trim()) nextErrors.contactName = 'Contact name is required'
    if (!form.email.trim()) nextErrors.email = 'Email is required'
    if (!form.region.trim()) nextErrors.region = 'Region is required'
    setFieldErrors(nextErrors)
    setMessage('')
    if (Object.keys(nextErrors).length > 0) return
    setPending(true)
    try {
      await updateProfile(form)
      onDone()
    } catch (error) {
      setFieldErrors(fieldErrorsOf(error))
      setMessage(isApiError(error) ? error.message : 'Profile could not be saved')
    } finally {
      setPending(false)
    }
  }

  return (
    <form className="grid gap-4" onSubmit={handleSubmit} noValidate>
      <h2 className="font-serif text-2xl text-ink-950">Edit profile</h2>
      {message && <p role="alert" className="text-sm text-danger">{message}</p>}
      <TextField id="businessName" label="Business name" autoComplete="organization" value={form.businessName} error={fieldErrors.businessName} onChange={(event) => update('businessName', event.target.value)} />
      <TextField id="contactName" label="Contact name" autoComplete="name" value={form.contactName} error={fieldErrors.contactName} onChange={(event) => update('contactName', event.target.value)} />
      <TextField id="email" label="Email" type="email" autoComplete="email" value={form.email} error={fieldErrors.email} onChange={(event) => update('email', event.target.value)} />
      <TextField id="region" label="Region" value={form.region} error={fieldErrors.region} onChange={(event) => update('region', event.target.value)} />
      <div className="flex flex-wrap gap-3">
        <Button type="submit" disabled={pending}>{pending ? 'Saving…' : 'Save'}</Button>
        <Button variant="secondary" onClick={onDone} disabled={pending}>Cancel</Button>
      </div>
    </form>
  )
}

function ProfileRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between gap-4">
      <dt className="text-ink-700">{label}</dt>
      <dd className="text-right font-semibold text-ink-900">{value}</dd>
    </div>
  )
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-md bg-ink-50 p-3">
      <dt className="text-sm text-ink-700">{label}</dt>
      <dd className="mt-1 font-serif text-3xl tabular-nums text-ink-950">{formatNumber(value)}</dd>
    </div>
  )
}

function initials(name: string) {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0].toUpperCase())
    .join('')
}
