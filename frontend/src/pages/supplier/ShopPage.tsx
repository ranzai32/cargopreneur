import { useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import type { FormEvent } from 'react'
import { Link, NavLink } from 'react-router-dom'
import { Button } from '../../components/Button'
import { Card } from '../../components/Card'
import { Container } from '../../components/Container'
import { PageHeader } from '../../components/PageHeader'
import { PageLayout } from '../../components/PageLayout'
import { TextField } from '../../components/TextField'
import { api } from '../../lib/api'
import { fieldErrorsOf, isApiError, useAppState } from '../../lib/app-state'
import { formatDeliveryStatus, formatNumber } from '../../lib/format'
import type { ApiOrder, ApiProduct } from '../../types/catalog'
import { StatusBadge } from '../OrdersPage'

const links = [
  { to: '/supplier', label: 'Products' },
  { to: '/supplier/orders', label: 'Orders' },
]

export function SupplierProductsPage() {
  return (
    <SupplierFrame title="Your catalog" lede="Add and edit the products buyers see under your supplier page.">
      <ProductDesk />
    </SupplierFrame>
  )
}

export function SupplierOrdersPage() {
  return (
    <SupplierFrame title="Your orders" lede="Ship a paid participation, then mark it delivered.">
      <OrderDesk />
    </SupplierFrame>
  )
}

function SupplierFrame({ title, lede, children }: { title: string; lede: string; children: ReactNode }) {
  const { user, ready } = useAppState()
  const supplier = user?.role === 'supplier' && user.supplierId
  return (
    <PageLayout>
      <Container className="py-section">
        <PageHeader title={title} lede={lede} />
        {!ready && <p className="mt-8 text-ink-700">Loading…</p>}
        {ready && !supplier && <p role="alert" className="mt-8 text-danger">You do not have access to this</p>}
        {ready && supplier && (
          <>
            <nav aria-label="Supplier" className="mt-6 flex flex-wrap gap-2">
              {links.map((link) => (
                <NavLink key={link.to} to={link.to} end={link.to === '/supplier'} className={({ isActive }) => isActive ? 'rounded-md bg-accent-700 px-3 py-2 text-sm font-semibold text-white' : 'rounded-md border border-ink-300 bg-white px-3 py-2 text-sm font-semibold text-ink-800'}>
                  {link.label}
                </NavLink>
              ))}
              <Link to={`/suppliers/${user.supplierId}`} className="rounded-md border border-ink-300 bg-white px-3 py-2 text-sm font-semibold text-ink-800">Public page</Link>
            </nav>
            <div className="mt-8">{children}</div>
          </>
        )}
      </Container>
    </PageLayout>
  )
}

function ProductDesk() {
  const [products, setProducts] = useState<ApiProduct[]>([])
  const [editing, setEditing] = useState<ApiProduct | null>(null)
  const [message, setMessage] = useState('')

  async function reload() {
    setProducts(await api.supplierProducts())
  }

  useEffect(() => {
    reload().catch((error) => setMessage(isApiError(error) ? error.message : 'Could not load products'))
  }, [])

  return (
    <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(18rem,24rem)]">
      <ProductForm
        product={editing}
        onSaved={async (text) => {
          setEditing(null)
          setMessage(text)
          await reload()
        }}
      />
      <div>
        {message && <p className="mb-4 text-sm text-ink-800">{message}</p>}
        <ul className="grid gap-3">
          {products.map((product) => (
            <li key={product.id}>
              <Card className="p-4">
                <h2 className="font-serif text-xl text-ink-950">{product.name}</h2>
                <p className="mt-1 text-sm text-ink-700">{product.category} · {product.reviewCount} reviews</p>
                <div className="mt-3 flex gap-2">
                  <Button size="sm" variant="secondary" onClick={() => setEditing(product)}>Edit</Button>
                  <Button size="sm" variant="ghost" onClick={() => void removeProduct(product, reload, setMessage)}>Delete</Button>
                </div>
              </Card>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}

async function removeProduct(product: ApiProduct, reload: () => Promise<void>, setMessage: (value: string) => void) {
  try {
    await api.deleteSupplierProduct(product.id)
    setMessage(`${product.name} was removed.`)
    await reload()
  } catch (error) {
    setMessage(isApiError(error) ? error.message : 'Could not delete the product')
  }
}

function ProductForm({ product, onSaved }: { product: ApiProduct | null; onSaved: (message: string) => Promise<void> }) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [category, setCategory] = useState('')
  const [tags, setTags] = useState('')
  const [unit, setUnit] = useState('')
  const [retailPrice, setRetailPrice] = useState('')
  const [minQty, setMinQty] = useState('10')
  const [unitPrice, setUnitPrice] = useState('')
  const [imageUrl, setImageUrl] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)

  useEffect(() => {
    setName(product?.name ?? '')
    setDescription(product?.description ?? '')
    setCategory(product?.category ?? '')
    setTags(product?.tags.join(', ') ?? '')
    setUnit(product?.unit ?? '')
    setRetailPrice(product ? String(product.retailPrice) : '')
    setMinQty(product?.priceTiers[0] ? String(product.priceTiers[0].minQty) : '10')
    setUnitPrice(product?.priceTiers[0] ? String(product.priceTiers[0].unitPrice) : '')
    setImageUrl(product?.imageUrl ?? '')
    setErrors({})
  }, [product])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setPending(true)
    setErrors({})
    const body = {
      supplierId: product?.supplier.id ?? '',
      name: name.trim(),
      description: description.trim(),
      category: category.trim(),
      tags: tags.split(',').map((tag) => tag.trim()).filter(Boolean),
      unit: unit.trim(),
      retailPrice: Number(retailPrice),
      imageUrl: imageUrl || undefined,
      priceTiers: [{ minQty: Number(minQty), unitPrice: Number(unitPrice) }],
    }
    try {
      if (product) await api.updateSupplierProduct(product.id, body)
      else await api.createSupplierProduct(body)
      await onSaved(product ? `${body.name} was updated.` : `${body.name} was added.`)
      if (!product) {
        setName('')
        setDescription('')
        setCategory('')
        setTags('')
        setUnit('')
        setRetailPrice('')
        setUnitPrice('')
        setImageUrl('')
      }
    } catch (error) {
      setErrors(fieldErrorsOf(error))
    } finally {
      setPending(false)
    }
  }

  return (
    <Card className="p-5 sm:p-6">
      <h2 className="font-serif text-2xl text-ink-950">{product ? 'Edit product' : 'New product'}</h2>
      <form className="mt-4 grid gap-4" onSubmit={handleSubmit}>
        <TextField id="sp-name" label="Name" value={name} error={errors.name} onChange={(event) => setName(event.target.value)} />
        <TextField id="sp-description" label="Description" value={description} error={errors.description} onChange={(event) => setDescription(event.target.value)} />
        <TextField id="sp-category" label="Category" value={category} error={errors.category} onChange={(event) => setCategory(event.target.value)} />
        <TextField id="sp-tags" label="Tags" value={tags} error={errors.tags} onChange={(event) => setTags(event.target.value)} />
        <TextField id="sp-unit" label="Unit" value={unit} error={errors.unit} onChange={(event) => setUnit(event.target.value)} />
        <TextField id="sp-retail" label="Retail price" type="number" min={0} step="0.01" value={retailPrice} error={errors.retailPrice} onChange={(event) => setRetailPrice(event.target.value)} />
        <TextField id="sp-min" label="Wholesale minimum" type="number" min={1} value={minQty} error={errors['priceTiers.0.minQty']} onChange={(event) => setMinQty(event.target.value)} />
        <TextField id="sp-price" label="Wholesale unit price" type="number" min={0} step="0.01" value={unitPrice} error={errors['priceTiers.0.unitPrice']} onChange={(event) => setUnitPrice(event.target.value)} />
        <label className="block text-sm font-semibold text-ink-900">
          Photo
          <input type="file" accept="image/jpeg,image/png,image/webp" className="mt-1.5 block w-full text-sm" onChange={(event) => void uploadPhoto(event.target.files?.[0], setImageUrl)} />
        </label>
        <Button type="submit" disabled={pending}>{pending ? 'Saving…' : product ? 'Save product' : 'Add product'}</Button>
      </form>
    </Card>
  )
}

async function uploadPhoto(file: File | undefined, setImageUrl: (value: string) => void) {
  if (!file) return
  setImageUrl(await api.supplierUpload(file))
}

function OrderDesk() {
  const [orders, setOrders] = useState<ApiOrder[]>([])
  const [message, setMessage] = useState('')
  const [updating, setUpdating] = useState('')

  async function reload() {
    setOrders(await api.supplierOrders())
  }

  useEffect(() => {
    reload().catch((error) => setMessage(isApiError(error) ? error.message : 'Could not load orders'))
  }, [])

  async function update(id: string, ship: boolean) {
    setUpdating(id)
    setMessage('')
    try {
      const updated = ship ? await api.supplierShip(id) : await api.supplierDeliver(id)
      setOrders((current) => current.map((item) => (item.id === updated.id ? updated : item)))
    } catch (error) {
      setMessage(isApiError(error) ? error.message : 'Could not update the shipment')
    } finally {
      setUpdating('')
    }
  }

  return (
    <div>
      {message && <p className="mb-4 text-sm text-danger">{message}</p>}
      <ul className="grid gap-4">
        {orders.map((order) => (
          <li key={order.id}>
            <Card className="p-5">
              <StatusBadge status={order.status} />
              <h2 className="mt-3 font-serif text-2xl text-ink-950">{order.product.name}</h2>
              <p className="mt-1 text-sm text-ink-700">{formatNumber(order.collectedQty)} / {formatNumber(order.moq)} {order.product.unit}</p>
              <ul className="mt-4 grid gap-3">
                {(order.participants ?? []).map((participant) => (
                  <li key={participant.id} className="rounded-md border border-ink-200 p-3">
                    <p className="text-sm text-ink-800">
                      {formatNumber(participant.quantity)} {order.product.unit}
                      {participant.shippingAddress ? ` · ${participant.shippingAddress}` : ''}
                      {` · ${participant.paymentStatus === 'paid' ? 'Paid' : 'Unpaid'}`}
                      {` · ${formatDeliveryStatus(participant.deliveryStatus)}`}
                    </p>
                    <div className="mt-3 flex gap-2">
                      <Button size="sm" disabled={updating !== '' || participant.paymentStatus !== 'paid' || participant.deliveryStatus !== 'pending'} onClick={() => void update(participant.id, true)}>Ship</Button>
                      <Button size="sm" variant="secondary" disabled={updating !== '' || participant.paymentStatus !== 'paid' || participant.deliveryStatus !== 'shipped'} onClick={() => void update(participant.id, false)}>Deliver</Button>
                    </div>
                  </li>
                ))}
              </ul>
            </Card>
          </li>
        ))}
      </ul>
    </div>
  )
}
