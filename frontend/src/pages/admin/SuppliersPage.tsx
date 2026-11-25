import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { Badge } from '../../components/Badge'
import { Button } from '../../components/Button'
import { Card } from '../../components/Card'
import { AddressField } from '../../components/AddressField'
import { TextField } from '../../components/TextField'
import { fieldErrorsOf, isApiError, useAppState } from '../../lib/app-state'
import { formatNumber } from '../../lib/format'
import type { ApiSupplier } from '../../types/catalog'
import { AdminFrame } from './AdminFrame'

const empty = { name: '', description: '', region: '' }

export function AdminSuppliersPage() {
  return (
    <AdminFrame
      title="Suppliers"
      lede="Add a supplier here. New suppliers stay out of the public catalog until you verify them."
    >
      <SupplierManager />
    </AdminFrame>
  )
}

function SupplierManager() {
  const { loadAdminSuppliers, createSupplier, verifySupplier } = useAppState()
  const [suppliers, setSuppliers] = useState<ApiSupplier[]>([])
  const [loaded, setLoaded] = useState(false)
  const [form, setForm] = useState(empty)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [message, setMessage] = useState('')
  const [problem, setProblem] = useState(false)
  const [pending, setPending] = useState(false)
  const [verifying, setVerifying] = useState('')
  const [blocked, setBlocked] = useState('')

  useEffect(() => {
    let cancel = false
    loadAdminSuppliers()
      .then((items) => {
        if (!cancel) {
          setSuppliers(items)
          setLoaded(true)
        }
      })
      .catch((error) => {
        if (cancel) return
        if (isApiError(error) && (error.status === 401 || error.status === 403)) {
          setBlocked(error.status === 403 ? 'You do not have access to this' : error.message)
          return
        }
        setProblem(true)
        setMessage(isApiError(error) ? error.message : 'Suppliers could not be loaded')
        setLoaded(true)
      })
    return () => {
      cancel = true
    }
  }, [loadAdminSuppliers])

  function update(key: keyof typeof empty, value: string) {
    setForm((current) => ({ ...current, [key]: value }))
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const nextErrors: Record<string, string> = {}
    if (!form.name.trim()) nextErrors.name = 'Name is required'
    if (!form.region.trim()) nextErrors.region = 'Region is required'
    setFieldErrors(nextErrors)
    setProblem(false)
    setMessage('')
    if (Object.keys(nextErrors).length > 0) return
    setPending(true)
    try {
      const created = await createSupplier({
        name: form.name.trim(),
        description: form.description.trim(),
        region: form.region.trim(),
      })
      setSuppliers((current) => [...current, created].sort((a, b) => a.name.localeCompare(b.name)))
      setForm(empty)
      setProblem(false)
      setMessage('Supplier added.')
    } catch (error) {
      if (isApiError(error) && (error.status === 401 || error.status === 403)) {
        setBlocked(error.status === 403 ? 'You do not have access to this' : error.message)
        return
      }
      setFieldErrors(fieldErrorsOf(error))
      setProblem(true)
      setMessage(isApiError(error) ? error.message : 'Could not add the supplier')
    } finally {
      setPending(false)
    }
  }

  async function handleVerify(id: string) {
    setVerifying(id)
    setProblem(false)
    setMessage('')
    try {
      const verified = await verifySupplier(id)
      setSuppliers((current) => current.map((item) => (item.id === id ? verified : item)))
      setProblem(false)
      setMessage('Supplier verified.')
    } catch (error) {
      if (isApiError(error) && (error.status === 401 || error.status === 403)) {
        setBlocked(error.status === 403 ? 'You do not have access to this' : error.message)
        return
      }
      setProblem(true)
      setMessage(isApiError(error) ? error.message : 'Could not verify the supplier')
    } finally {
      setVerifying('')
    }
  }

  if (blocked) {
    return <p role="alert" className="text-danger">{blocked}</p>
  }

  return (
    <div className="grid gap-8 lg:grid-cols-[minmax(0,22rem)_1fr] lg:items-start">
      <Card className="p-6">
        <h2 className="font-serif text-2xl text-ink-950">Add a supplier</h2>
        <form className="mt-4 grid gap-4" onSubmit={handleSubmit} noValidate>
          {message && (
            <p role={problem ? 'alert' : 'status'} className={problem ? 'text-sm text-danger' : 'text-sm text-ink-800'}>
              {message}
            </p>
          )}
          <TextField id="supplier-name" label="Name" value={form.name} error={fieldErrors.name} onChange={(event) => update('name', event.target.value)} />
          <TextField id="supplier-description" label="Description" value={form.description} error={fieldErrors.description} onChange={(event) => update('description', event.target.value)} />
          <AddressField id="supplier-region" label="Region" value={form.region} error={fieldErrors.region} onChange={(value) => update('region', value)} />
          <Button type="submit" disabled={pending}>{pending ? 'Adding…' : 'Add supplier'}</Button>
        </form>
      </Card>
      <div>
        {!loaded && <p className="text-ink-700">Loading suppliers…</p>}
        {loaded && suppliers.length === 0 && !message && <p className="text-ink-700">No suppliers yet.</p>}
        <ul className="grid gap-4">
          {suppliers.map((supplier) => {
            const count = supplier.productCount ?? 0
            const countLabel = count === 1 ? '1 product' : `${formatNumber(count)} products`
            return (
              <li key={supplier.id}>
                <Card className="p-5">
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge tone={supplier.isVerified ? 'success' : 'warning'}>
                      {supplier.isVerified ? 'Verified' : 'Not verified'}
                    </Badge>
                    <span className="text-sm text-ink-700">{countLabel}</span>
                  </div>
                  <h3 className="mt-3 font-serif text-2xl text-ink-950">{supplier.name}</h3>
                  <p className="mt-1 text-sm text-ink-700">{supplier.region}</p>
                  {supplier.description && <p className="mt-3 text-sm text-ink-800">{supplier.description}</p>}
                  {!supplier.isVerified && (
                    <div className="mt-4">
                      <Button size="sm" variant="secondary" disabled={verifying === supplier.id} onClick={() => void handleVerify(supplier.id)}>
                        {verifying === supplier.id ? 'Verifying…' : 'Verify'}
                      </Button>
                    </div>
                  )}
                </Card>
              </li>
            )
          })}
        </ul>
      </div>
    </div>
  )
}
