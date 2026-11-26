import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Badge } from '../components/Badge'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageLayout } from '../components/PageLayout'
import { RatingLink, ReviewForm, ReviewList } from '../components/Reviews'
import { api } from '../lib/api'
import { isApiError, useAppState } from '../lib/app-state'
import { mediaUrl } from '../lib/media'
import type { ApiProduct, ApiReview, ApiSupplier } from '../types/catalog'

export function SupplierPage() {
  const { supplierId = '' } = useParams()
  const { user, loadProducts } = useAppState()
  const [supplier, setSupplier] = useState<ApiSupplier | null>(null)
  const [products, setProducts] = useState<ApiProduct[]>([])
  const [reviews, setReviews] = useState<ApiReview[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    let cancel = false
    api.supplier(supplierId)
      .then((found) => {
        if (cancel) return
        setSupplier(found)
        document.title = `${found.name} · CargoPrener`
        return Promise.all([
          api.supplierReviews(supplierId).catch(() => [] as ApiReview[]),
          loadProducts({ supplierId, pageSize: 50, sort: 'rating' }).catch(() => ({ items: [] as ApiProduct[] })),
        ])
      })
      .then((extra) => {
        if (cancel || !extra) return
        setReviews(extra[0])
        setProducts(extra[1].items)
      })
      .catch((reason) => {
        if (!cancel) setError(isApiError(reason) ? reason.message : 'Supplier not found')
      })
    return () => {
      cancel = true
    }
  }, [loadProducts, supplierId])

  async function reloadReviews() {
    setReviews(await api.supplierReviews(supplierId))
    setSupplier(await api.supplier(supplierId))
  }

  return (
    <PageLayout>
      <Container className="py-section">
        <Link to="/catalog" className="text-sm font-semibold text-accent-800">Back to catalog</Link>
        {error && <p role="alert" className="mt-6 text-danger">{error}</p>}
        {supplier && (
          <div className="mt-6 grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(18rem,24rem)]">
            <div>
              <Badge tone="accent">{supplier.isVerified ? 'Reviewed supplier' : 'Under review'}</Badge>
              <h1 className="mt-3 font-serif text-4xl text-ink-950">{supplier.name}</h1>
              <p className="mt-2 text-sm text-ink-700">{supplier.region}</p>
              <p className="mt-2">
                <RatingLink to={`/suppliers/${supplier.id}/reviews`} rating={supplier.rating} count={supplier.reviewCount} />
              </p>
              {supplier.description && <p className="mt-4 max-w-2xl text-ink-800">{supplier.description}</p>}
              <h2 className="mt-8 font-serif text-2xl text-ink-950">Catalog</h2>
              <ul className="mt-4 grid gap-4 sm:grid-cols-2">
                {products.map((product) => (
                  <li key={product.id}>
                    <Card className="h-full p-4">
                      <Link to={`/catalog/${product.id}`} className="block">
                        {product.imageUrl && <img src={mediaUrl(product.imageUrl)} alt="" className="mb-3 aspect-[4/3] w-full rounded-md object-cover" />}
                        <h3 className="font-serif text-xl text-ink-950">{product.name}</h3>
                      </Link>
                      <p className="mt-1">
                        <RatingLink to={`/catalog/${product.id}/reviews`} rating={product.rating} count={product.reviewCount} />
                      </p>
                    </Card>
                  </li>
                ))}
              </ul>
            </div>
            <Card className="h-fit p-5 sm:p-6">
              <h2 className="font-serif text-2xl text-ink-950">Reviews</h2>
              <ReviewList reviews={reviews} />
              {user && user.supplierId !== supplier.id && (
                <ReviewForm onSubmit={async (rating, body) => {
                  await api.writeSupplierReview(supplier.id, { rating, body })
                  await reloadReviews()
                }} />
              )}
            </Card>
          </div>
        )}
      </Container>
    </PageLayout>
  )
}

