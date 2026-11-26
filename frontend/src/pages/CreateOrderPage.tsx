import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button } from '../components/Button'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageHeader } from '../components/PageHeader'
import { PageLayout } from '../components/PageLayout'
import { SelectField } from '../components/TextField'
import { TextField } from '../components/TextField'
import { fieldErrorsOf, isApiError, useAppState } from '../lib/app-state'
import { formatMoney } from '../lib/format'
import { mediaUrl } from '../lib/media'
import type { ApiProduct, ProductPage } from '../types/catalog'

export function CreateOrderPage() {
  const { user, loadProducts, createOrder } = useAppState()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const requestedProduct = params.get('product') ?? ''
  const [products, setProducts] = useState<ApiProduct[]>([])
  const [supplierId, setSupplierId] = useState('')
  const [productId, setProductId] = useState(requestedProduct)
  const [moq, setMoq] = useState('100')
  const [deadline, setDeadline] = useState(defaultDeadline)
  const [quantity, setQuantity] = useState('10')
  const [message, setMessage] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)

  useEffect(() => {
    document.title = 'Create an order · CargoPrener'
    let cancel = false
    loadCatalog(loadProducts)
      .then((items) => {
        if (cancel) return
        const requested = items.find((item) => item.id === requestedProduct)
        const nextSupplier = requested?.supplier.id ?? suppliersFrom(items)[0]?.id ?? ''
        const nextProduct = requested?.id ?? productsForSupplier(items, nextSupplier)[0]?.id ?? ''
        setProducts(items)
        setSupplierId(nextSupplier)
        setProductId(nextProduct)
      })
      .catch(() => {
        if (!cancel) setMessage('The catalog could not be loaded')
      })
    return () => {
      cancel = true
    }
  }, [loadProducts, requestedProduct])

  const suppliers = useMemo(() => suppliersFrom(products), [products])
  const supplierProducts = useMemo(
    () => productsForSupplier(products, supplierId),
    [products, supplierId],
  )
  const selected = supplierProducts.find((item) => item.id === productId)

  useEffect(() => {
    const selected = supplierProducts.find((item) => item.id === productId)
    const firstTier = selected?.priceTiers[0]
    if (firstTier) setMoq(String(firstTier.minQty))
  }, [productId, supplierProducts])

  function handleSupplierChange(nextSupplierId: string) {
    const nextProducts = productsForSupplier(products, nextSupplierId)
    setSupplierId(nextSupplierId)
    setProductId(nextProducts[0]?.id ?? '')
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const nextErrors: Record<string, string> = {}
    if (!supplierId) nextErrors.supplierId = 'Choose a supplier'
    if (!supplierProducts.some((item) => item.id === productId)) {
      nextErrors.productId = 'Choose a product from this supplier'
    }
    if (!Number(moq) || Number(moq) < 1) nextErrors.moq = 'Enter a minimum of at least 1'
    if (!deadline) nextErrors.deadline = 'Choose a deadline'
    if (!Number(quantity) || Number(quantity) < 1) nextErrors.quantity = 'Enter the quantity you need'
    setFieldErrors(nextErrors)
    setMessage('')
    if (Object.keys(nextErrors).length > 0) return
    setPending(true)
    try {
      const created = await createOrder({
        productId,
        moq: Number(moq),
        deadline,
        quantity: Number(quantity),
      })
      navigate(`/orders/${created.id}`)
    } catch (error) {
      setFieldErrors(fieldErrorsOf(error))
      setMessage(isApiError(error) ? error.message : 'Could not create the order')
    } finally {
      setPending(false)
    }
  }

  return (
    <PageLayout>
      <Container className="py-section">
        <PageHeader
          title="Start a pooled order"
          lede="You set the product, the minimum, and your own quantity. The delivery address is the next step, and only after the pool reaches that minimum. OpenStreetMap suggests the street while you type."
        />
        {!user ? (
          <Card className="mt-8 max-w-xl p-6">
            <p className="text-ink-800">Sign in before you start an order.</p>
            <div className="mt-4 flex gap-3">
              <Button to="/login">Log in</Button>
              <Button to="/register" variant="secondary">Create an account</Button>
            </div>
          </Card>
        ) : (
          <div className="mt-8 grid items-start gap-6 lg:grid-cols-[minmax(0,24rem)_minmax(0,1fr)]">
          <Card className="p-6">
            <form className="grid gap-4" onSubmit={handleSubmit} noValidate>
              {message && <p role="alert" className="text-sm text-danger">{message}</p>}
              <SelectField id="supplier" label="Supplier" value={supplierId} error={fieldErrors.supplierId} onChange={(event) => handleSupplierChange(event.target.value)}>
                {suppliers.length === 0 && <option value="">No suppliers yet</option>}
                {suppliers.map((supplier) => (
                  <option key={supplier.id} value={supplier.id}>{supplier.name}</option>
                ))}
              </SelectField>
              <SelectField id="product" label="Product" value={productId} error={fieldErrors.productId} onChange={(event) => setProductId(event.target.value)}>
                {supplierProducts.length === 0 && <option value="">No products for this supplier</option>}
                {supplierProducts.map((item) => (
                  <option key={item.id} value={item.id}>{item.name}</option>
                ))}
              </SelectField>
              <TextField id="moq" label="Supplier minimum" type="number" min={1} value={moq} error={fieldErrors.moq} onChange={(event) => setMoq(event.target.value)} />
              <TextField id="deadline" label="Deadline" type="date" value={deadline} error={fieldErrors.deadline} onChange={(event) => setDeadline(event.target.value)} />
              <TextField id="quantity" label="Your quantity" type="number" min={1} value={quantity} error={fieldErrors.quantity} onChange={(event) => setQuantity(event.target.value)} />
              <Button type="submit" disabled={pending}>{pending ? 'Creating…' : 'Create order'}</Button>
            </form>
          </Card>
          <Card className="p-6">
            {selected?.imageUrl && (
              <img src={mediaUrl(selected.imageUrl)} alt="" className="aspect-[4/3] w-full rounded-xl object-cover" />
            )}
            <h2 className="mt-4 font-serif text-2xl text-ink-950">{selected?.name ?? 'Choose a product'}</h2>
            <p className="mt-2 text-sm text-ink-700">{selected ? `${selected.supplier.name} · ${selected.supplier.region}` : 'The supplier list loads from the catalog.'}</p>
            {selected && (
              <ul className="mt-4 grid gap-2 text-sm text-ink-800">
                {selected.priceTiers.map((tier) => (
                  <li key={tier.minQty}>{tier.minQty}+ {selected.unit} at {formatMoney(tier.unitPrice)}</li>
                ))}
              </ul>
            )}
            <ol className="mt-6 grid gap-3 text-sm text-ink-800">
              <li>1. This page only records the pool: product, minimum, deadline, and your quantity.</li>
              <li>2. Other businesses can join while the order stays open.</li>
              <li>3. After the minimum is reached, checkout is a separate screen: line items, a saved address, a delivery method, and the total.</li>
              <li>4. The supplier marks the shipment and the delivery from their own account.</li>
            </ol>
          </Card>
          </div>
        )}
      </Container>
    </PageLayout>
  )
}

function defaultDeadline() {
  const date = new Date()
  date.setDate(date.getDate() + 14)
  return date.toISOString().slice(0, 10)
}

async function loadCatalog(loadProducts: (filters: { page: number; pageSize: number; sort: 'name' }) => Promise<ProductPage>) {
  const pageSize = 50
  const first = await loadProducts({ page: 1, pageSize, sort: 'name' })
  const items = [...first.items]
  const pages = Math.max(1, Math.ceil(first.total / first.pageSize))
  for (let page = 2; page <= pages; page += 1) {
    const next = await loadProducts({ page, pageSize, sort: 'name' })
    items.push(...next.items)
  }
  return items
}

function suppliersFrom(products: ApiProduct[]) {
  const seen = new Map<string, ApiProduct['supplier']>()
  for (const product of products) seen.set(product.supplier.id, product.supplier)
  return [...seen.values()].sort((left, right) => left.name.localeCompare(right.name))
}

function productsForSupplier(products: ApiProduct[], supplierId: string) {
  return products
    .filter((product) => product.supplier.id === supplierId)
    .sort((left, right) => left.name.localeCompare(right.name))
}
