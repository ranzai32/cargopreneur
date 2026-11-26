import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Badge } from '../components/Badge'
import { Button } from '../components/Button'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageLayout } from '../components/PageLayout'
import { isApiError, useAppState } from '../lib/app-state'
import { formatMoney, formatQuantity } from '../lib/format'
import { RatingLink, ReviewForm, ReviewList } from '../components/Reviews'
import { api } from '../lib/api'
import { mediaUrl } from '../lib/media'
import type { ApiProduct, ApiReview } from '../types/catalog'

export function ProductPage() {
  const { productId = '' } = useParams()
  const { user, loadProduct, loadOrders } = useAppState()
  const [product, setProduct] = useState<ApiProduct | null>(null)
  const [reviews, setReviews] = useState<ApiReview[]>([])
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
    api.productReviews(productId).then((items) => {
      if (!cancel) setReviews(items)
    }).catch(() => {
      if (!cancel) setReviews([])
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
              {product.imageUrl ? (
                <img src={mediaUrl(product.imageUrl)} alt="" className="mb-6 aspect-[4/3] w-full rounded-xl object-cover" />
              ) : (
                <div className="mb-6 flex aspect-[4/3] items-center justify-center rounded-xl bg-ink-100 text-sm font-semibold text-ink-600">{product.category}</div>
              )}
              <Badge tone="accent">{product.category}</Badge>
              <h1 className="mt-3 font-serif text-4xl leading-tight text-ink-950">{product.name}</h1>
              <p className="mt-3 text-ink-700">{product.description}</p>
              <p className="mt-4 text-sm text-ink-700">
                <Link to={`/suppliers/${product.supplier.id}`} className="font-semibold text-accent-800">{product.supplier.name}</Link>
                {' · '}{product.supplier.region}
                {product.supplier.isVerified ? ' · Reviewed supplier' : ''}
              </p>
              <p className="mt-2">
                <RatingLink to={`/catalog/${product.id}/reviews`} rating={product.rating} count={product.reviewCount} />
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
        {product && (
          <Card className="mt-8 p-5 sm:p-6">
            <h2 className="font-serif text-2xl text-ink-950">Reviews</h2>
            <ReviewList reviews={reviews} />
            {user && user.supplierId !== product.supplier.id && (
              <ReviewForm onSubmit={async (rating, body) => {
                await api.writeProductReview(product.id, { rating, body })
                setReviews(await api.productReviews(product.id))
                setProduct(await loadProduct(product.id))
              }} />
            )}
          </Card>
        )}
      </Container>
    </PageLayout>
  )
}
