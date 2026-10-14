import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Badge } from '../components/Badge'
import { Button } from '../components/Button'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageLayout } from '../components/PageLayout'
import { isApiError, useAppState } from '../lib/app-state'
import { formatMoney, formatQuantity } from '../lib/format'
import type { ApiProduct } from '../types/catalog'

export function ProductPage() {
  const { productId = '' } = useParams()
  const { loadProduct, loadOrders } = useAppState()
  const [product, setProduct] = useState<ApiProduct | null>(null)
  const [orderCount, setOrderCount] = useState(0)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancel = false
    loadProduct(productId)
      .then((item) => {
        if (cancel) return
        setProduct(item)
        document.title = `${item.name} · CargoPrener`
      })
      .catch((reason) => {
        if (!cancel) setError(isApiError(reason) ? reason.message : 'Product not found')
      })
    loadOrders('open')
      .then((orders) => {
        if (!cancel) setOrderCount(orders.filter((order) => order.product.id === productId).length)
      })
      .catch(() => {
        if (!cancel) setOrderCount(0)
      })
    return () => {
      cancel = true
    }
  }, [loadOrders, loadProduct, productId])

  return (
    <PageLayout>
      <Container className="py-section">
        <Link to="/catalog" className="text-sm font-semibold text-accent-800">Back to catalog</Link>
        {error && <p role="alert" className="mt-6 text-danger">{error}</p>}
        {product && (
          <div className="mt-6 grid gap-8 lg:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)]">
            <div>
              <Badge tone="accent">{product.category}</Badge>
              <h1 className="mt-3 font-serif text-4xl leading-tight text-ink-950">{product.name}</h1>
              <p className="mt-3 text-ink-700">{product.description}</p>
              <p className="mt-4 text-sm text-ink-700">
                {product.supplier.name} · {product.supplier.region}
                {product.supplier.isVerified ? ' · Reviewed supplier' : ''}
              </p>
              <p className="mt-6 text-ink-800">
                Retail price {formatMoney(product.retailPrice)} per {product.unit}. Wholesale applies only if a pooled order reaches the minimum.
              </p>
              <div className="mt-6 flex flex-col gap-3 sm:flex-row">
                <Button to={`/orders/new?product=${product.id}`}>Create an order</Button>
                <Button to="/orders" variant="secondary">See open orders</Button>
              </div>
              {orderCount > 0 && (
                <p className="mt-4 text-sm text-ink-700">
                  {orderCount} open {orderCount === 1 ? 'order' : 'orders'} already collecting this product.
                </p>
              )}
            </div>
            <Card className="p-5 sm:p-6">
              <h2 className="font-serif text-2xl text-ink-950">Wholesale tiers</h2>
              <ul className="mt-4 border-t border-ink-200">
                {product.priceTiers.map((tier) => (
                  <li key={tier.minQty} className="flex items-baseline justify-between gap-4 border-b border-ink-200 py-3 text-sm">
                    <span className="font-semibold text-ink-950">{formatQuantity(tier.minQty, product.unit)}+</span>
                    <span className="tabular-nums text-ink-800">{formatMoney(tier.unitPrice)} / {product.unit}</span>
                  </li>
                ))}
              </ul>
              <p className="mt-3 text-sm text-ink-700">Retail {formatMoney(product.retailPrice)} / {product.unit}</p>
            </Card>
          </div>
        )}
      </Container>
    </PageLayout>
  )
}
