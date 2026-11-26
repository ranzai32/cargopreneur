import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Container } from '../components/Container'
import { PageLayout } from '../components/PageLayout'
import { ReviewForm, ReviewList, Stars } from '../components/Reviews'
import { api } from '../lib/api'
import { useAppState } from '../lib/app-state'
import type { ApiReview } from '../types/catalog'

export function ProductReviewsPage() {
  const { productId = '' } = useParams()
  return <ReviewsScreen productId={productId} />
}

export function SupplierReviewsPage() {
  const { supplierId = '' } = useParams()
  return <ReviewsScreen supplierId={supplierId} />
}

function ReviewsScreen({ productId, supplierId }: { productId?: string; supplierId?: string }) {
  const { user } = useAppState()
  const [title, setTitle] = useState('')
  const [reviews, setReviews] = useState<ApiReview[]>([])
  const [ownerId, setOwnerId] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    let cancel = false
    const load = productId
      ? api.product(productId).then(async (product) => ({
          name: product.name,
          owner: product.supplier.id,
          reviews: await api.productReviews(productId),
        }))
      : api.supplier(supplierId ?? '').then(async (supplier) => ({
          name: supplier.name,
          owner: supplier.id,
          reviews: await api.supplierReviews(supplier.id),
        }))
    load.then((result) => {
      if (cancel) return
      setTitle(result.name)
      setOwnerId(result.owner)
      setReviews(result.reviews)
      document.title = `Reviews · ${result.name}`
    }).catch(() => {
      if (!cancel) setError('Reviews could not be loaded')
    })
    return () => {
      cancel = true
    }
  }, [productId, supplierId])

  const average = reviews.length
    ? reviews.reduce((sum, review) => sum + review.rating, 0) / reviews.length
    : 0

  async function reload() {
    setReviews(productId ? await api.productReviews(productId) : await api.supplierReviews(supplierId ?? ''))
  }

  return (
    <PageLayout>
      <Container className="py-section">
        <Link to={productId ? `/catalog/${productId}` : `/suppliers/${supplierId}`} className="text-sm font-semibold text-accent-800">
          {productId ? 'Back to product' : 'Back to supplier'}
        </Link>
        {error && <p role="alert" className="mt-6 text-danger">{error}</p>}
        {!error && (
          <>
            <h1 className="mt-4 font-serif text-4xl text-ink-950">{title || 'Reviews'}</h1>
            <p className="mt-3 flex items-center gap-2 text-sm text-ink-800">
              <Stars value={average} />
              <span>{reviews.length ? `${average.toFixed(1)} · ${reviews.length}` : 'No reviews yet'}</span>
            </p>
            <ReviewList reviews={reviews} />
            {user && user.supplierId !== ownerId && (
              <ReviewForm onSubmit={async (rating, body) => {
                if (productId) await api.writeProductReview(productId, { rating, body })
                else await api.writeSupplierReview(supplierId ?? '', { rating, body })
                await reload()
              }} />
            )}
          </>
        )}
      </Container>
    </PageLayout>
  )
}
