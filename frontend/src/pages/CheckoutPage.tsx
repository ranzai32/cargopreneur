import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { AddressField } from '../components/AddressField'
import { Button } from '../components/Button'
import { CheckoutFrame } from '../components/CheckoutFrame'
import { Container } from '../components/Container'
import { fieldErrorsOf, isApiError, useAppState } from '../lib/app-state'
import { createAddress, loadAddresses, removeAddress, updateAddress } from '../lib/addresses'
import type { SavedAddress } from '../lib/api'
import { deliveryChoices } from '../lib/delivery'
import { formatMoney, formatQuantity } from '../lib/format'
import { mediaUrl } from '../lib/media'
import { quoteQuantity } from '../lib/pricing'
import type { ApiOrder } from '../types/catalog'
import type { DeliveryChoice } from '../lib/delivery'

export function CheckoutPage() {
  const { orderId = '' } = useParams()
  const navigate = useNavigate()
  const { user, loadOrder, pay } = useAppState()
  const [order, setOrder] = useState<ApiOrder | null>(null)
  const [addresses, setAddresses] = useState<SavedAddress[]>([])
  const [selectedId, setSelectedId] = useState('')
  const [method, setMethod] = useState<DeliveryChoice['id']>('freight')
  const [editor, setEditor] = useState<AddressDraft | null>(null)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [pending, setPending] = useState(false)

  useEffect(() => {
    document.title = 'Checkout · CargoPrener'
    let cancel = false
    loadOrder(orderId)
      .then((item) => {
        if (!cancel) setOrder(item)
      })
      .catch((reason) => {
        if (!cancel) setError(isApiError(reason) ? reason.message : 'Order not found')
      })
    return () => {
      cancel = true
    }
  }, [loadOrder, orderId])

  useEffect(() => {
    if (!user) return
    let cancel = false
    loadAddresses(user.id)
      .then((items) => {
        if (cancel) return
        setAddresses(items)
        setSelectedId((current) => current || items[0]?.id || '')
      })
      .catch((reason) => {
        if (!cancel) setMessage(isApiError(reason) ? reason.message : 'Addresses could not be loaded')
      })
    return () => {
      cancel = true
    }
  }, [user])

  if (error || !order) {
    return (
      <CheckoutFrame backTo="/orders" backLabel="Orders">
        <Container>
          <p role="alert" className="text-danger">{error || 'Loading checkout…'}</p>
        </Container>
      </CheckoutFrame>
    )
  }

  const quantity = order.yourQuantity ?? 0
  const payable = Boolean(user && quantity && order.status === 'confirmed' && order.yourPaymentStatus !== 'paid')
  const choices = deliveryChoices(quantity || 1)
  const delivery = choices.find((item) => item.id === method) ?? choices[1]
  const quote = quoteQuantity({
    collectedQty: Math.max(0, order.collectedQty - quantity),
    yourQty: quantity,
    minimumQty: order.moq,
    retailPrice: order.product.retailPrice,
    tiers: order.product.priceTiers.map((tier) => ({ minQty: tier.minQty, price: tier.unitPrice })),
  })
  const goods = quote.yourTotal
  const total = goods === null ? null : goods + delivery.price
  const selected = addresses.find((item) => item.id === selectedId) ?? null

  async function handleSaveAddress(draft: AddressDraft) {
    if (!user) return
    setPending(true)
    setMessage('')
    try {
      const saved = draft.id
        ? await updateAddress(user.id, draft.id, draft)
        : await createAddress(user.id, draft)
      setAddresses((current) => [saved, ...current.filter((item) => item.id !== saved.id)])
      setSelectedId(saved.id)
      setEditor(null)
    } catch (reason) {
      const fields = fieldErrorsOf(reason)
      setEditor((current) => current && { ...current, errors: fields })
      if (!fields.line && !fields.label) setMessage(isApiError(reason) ? reason.message : 'Could not save the address')
    } finally {
      setPending(false)
    }
  }

  async function handleRemove(id: string) {
    if (!user) return
    setPending(true)
    setMessage('')
    try {
      await removeAddress(user.id, id)
      const next = addresses.filter((item) => item.id !== id)
      setAddresses(next)
      setSelectedId((current) => (current === id ? next[0]?.id ?? '' : current))
      setEditor((current) => (current?.id === id ? null : current))
    } catch (reason) {
      setMessage(isApiError(reason) ? reason.message : 'Could not remove the address')
    } finally {
      setPending(false)
    }
  }

  async function handlePlaceOrder() {
    if (!selected) {
      setMessage('Choose a delivery address, or save a new one.')
      return
    }
    setPending(true)
    setMessage('')
    try {
      await pay(orderId, { shippingAddress: selected.line, deliveryMethod: delivery.id })
      navigate(`/orders/${orderId}/processing`, { replace: true })
    } catch (reason) {
      setMessage(isApiError(reason) ? reason.message : 'Could not place the order')
      setPending(false)
    }
  }

  return (
    <CheckoutFrame backTo={`/orders/${order.id}`} backLabel="Back to the order">
      <Container>
        {!user && (
          <p className="text-ink-800">
            <Link to="/login" state={{ from: `/orders/${order.id}/checkout` }} className="font-semibold text-accent-800">Log in</Link>
            {' '}to check out.
          </p>
        )}
        {user && !payable && (
          <p className="text-ink-800">
            {order.yourPaymentStatus === 'paid'
              ? 'This order is already paid.'
              : 'Checkout opens after you join and the pool reaches its minimum.'}
            {' '}
            <Link to={`/orders/${order.id}`} className="font-semibold text-accent-800">Return to the order</Link>
          </p>
        )}
        {user && payable && (
          <>
            <div className="flex items-center gap-4">
              {order.product.imageUrl && (
                <img src={mediaUrl(order.product.imageUrl)} alt="" className="size-24 rounded-xl object-cover" />
              )}
              <div>
                <p className="text-sm font-semibold uppercase tracking-wide text-ink-600">Checkout</p>
                <h1 className="font-serif text-3xl text-ink-950">{order.product.name}</h1>
                <p className="text-sm text-ink-700">{order.product.supplier.name} · {formatQuantity(quantity, order.product.unit)}</p>
              </div>
            </div>
            <div className="mt-8 grid items-start gap-8 lg:grid-cols-[minmax(0,1fr)_22rem]">
              <div className="grid gap-8">
                <section>
                  <h2 className="font-serif text-2xl text-ink-950">Delivery</h2>
                  <div className="mt-3 grid gap-2" role="radiogroup" aria-label="Delivery method">
                    {choices.map((item) => (
                      <button
                        key={item.id}
                        type="button"
                        role="radio"
                        aria-checked={item.id === delivery.id}
                        className={item.id === delivery.id ? 'flex items-start justify-between gap-4 rounded-lg border border-accent-700 bg-white px-4 py-3 text-left' : 'flex items-start justify-between gap-4 rounded-lg border border-ink-200 bg-ink-50 px-4 py-3 text-left'}
                        onClick={() => setMethod(item.id)}
                      >
                        <span>
                          <span className="block font-semibold text-ink-950">{item.name}</span>
                          <span className="mt-0.5 block text-sm text-ink-700">{item.detail}</span>
                        </span>
                        <span className="font-semibold tabular-nums text-ink-950">{item.price === 0 ? 'Free' : formatMoney(item.price)}</span>
                      </button>
                    ))}
                  </div>
                </section>
                <section>
                  <div className="flex items-end justify-between gap-3">
                    <h2 className="font-serif text-2xl text-ink-950">Address</h2>
                    {!editor && (
                      <button type="button" className="text-sm font-semibold text-accent-800" onClick={() => setEditor({ label: '', line: '' })}>
                        Add an address
                      </button>
                    )}
                  </div>
                  <p className="mt-1 text-sm text-ink-700">Saved addresses stay on your account. Pick one for this shipment, or edit it.</p>
                  {editor && (
                    <AddressEditor
                      draft={editor}
                      pending={pending}
                      onChange={setEditor}
                      onCancel={() => setEditor(null)}
                      onSave={() => void handleSaveAddress(editor)}
                    />
                  )}
                  <ul className="mt-3 grid gap-2">
                    {addresses.length === 0 && !editor && <li className="text-sm text-ink-700">No saved addresses yet.</li>}
                    {addresses.map((item) => (
                      <li key={item.id}>
                        <div className={item.id === selectedId ? 'rounded-lg border border-accent-700 bg-white px-4 py-3' : 'rounded-lg border border-ink-200 bg-ink-50 px-4 py-3'}>
                          <button type="button" className="w-full text-left" onClick={() => setSelectedId(item.id)}>
                            <span className="block font-semibold text-ink-950">{item.label}</span>
                            <span className="mt-0.5 block text-sm text-ink-700">{item.line}</span>
                          </button>
                          <div className="mt-2 flex gap-3 text-sm font-semibold">
                            <button type="button" className="text-accent-800" onClick={() => setEditor({ id: item.id, label: item.label, line: item.line })}>Edit</button>
                            <button type="button" className="text-ink-700" onClick={() => void handleRemove(item.id)}>Remove</button>
                          </div>
                        </div>
                      </li>
                    ))}
                  </ul>
                </section>
              </div>
              <aside className="rounded-xl border border-ink-200 bg-white p-5 shadow-card lg:sticky lg:top-6">
                <h2 className="font-serif text-2xl text-ink-950">Your order</h2>
                <dl className="mt-4 grid gap-3 text-sm">
                  <div className="flex items-start justify-between gap-4">
                    <dt>
                      <span className="font-semibold text-ink-950">{order.product.name}</span>
                      <span className="mt-0.5 block text-ink-700">
                        {formatQuantity(quantity, order.product.unit)}
                        {quote.unitPrice !== null && ` × ${formatMoney(quote.unitPrice)}`}
                      </span>
                    </dt>
                    <dd className="font-semibold tabular-nums text-ink-950">{goods === null ? '—' : formatMoney(goods)}</dd>
                  </div>
                  <div className="flex items-start justify-between gap-4">
                    <dt>
                      <span className="font-semibold text-ink-950">{delivery.name}</span>
                      <span className="mt-0.5 block text-ink-700">{selected ? selected.line : 'No address selected'}</span>
                    </dt>
                    <dd className="font-semibold tabular-nums text-ink-950">{delivery.price === 0 ? 'Free' : formatMoney(delivery.price)}</dd>
                  </div>
                </dl>
                <div className="mt-4 flex items-end justify-between border-t border-ink-200 pt-4">
                  <span className="text-sm font-semibold uppercase tracking-wide text-ink-600">Total</span>
                  <span className="font-serif text-3xl tabular-nums text-ink-950">{total === null ? '—' : formatMoney(total)}</span>
                </div>
                <p className="mt-3 text-sm text-ink-700">
                  {total === null ? 'This quantity is below the first wholesale price.' : 'Simulated payment. No card is charged.'}
                </p>
                {message && <p role="alert" className="mt-3 text-sm text-danger">{message}</p>}
                <Button className="mt-4 w-full" size="lg" disabled={pending || !selected || total === null} onClick={() => void handlePlaceOrder()}>
                  {pending ? 'Placing order…' : 'Place order'}
                </Button>
              </aside>
            </div>
          </>
        )}
      </Container>
    </CheckoutFrame>
  )
}

type AddressDraft = {
  id?: string
  label: string
  line: string
  errors?: Record<string, string>
}

function AddressEditor({
  draft,
  pending,
  onChange,
  onCancel,
  onSave,
}: {
  draft: AddressDraft
  pending: boolean
  onChange: (draft: AddressDraft) => void
  onCancel: () => void
  onSave: () => void
}) {
  return (
    <div className="mt-3 grid gap-3 rounded-lg border border-ink-200 bg-white p-4">
      <label className="block" htmlFor="address-label">
        <span className="text-sm font-semibold text-ink-900">Name</span>
        <input
          id="address-label"
          value={draft.label}
          placeholder="Shop, warehouse, home"
          className="mt-1.5 h-11 w-full rounded-md border border-ink-300 px-3"
          onChange={(event) => onChange({ ...draft, label: event.target.value, errors: undefined })}
        />
        {draft.errors?.label && <span className="mt-1 block text-sm text-danger">{draft.errors.label}</span>}
      </label>
      <AddressField
        id="address-line"
        label="Street address"
        value={draft.line}
        error={draft.errors?.line}
        onChange={(line) => onChange({ ...draft, line, errors: undefined })}
      />
      <div className="flex gap-3">
        <Button disabled={pending} onClick={onSave}>{draft.id ? 'Save changes' : 'Save address'}</Button>
        <Button variant="secondary" onClick={onCancel}>Cancel</Button>
      </div>
    </div>
  )
}
