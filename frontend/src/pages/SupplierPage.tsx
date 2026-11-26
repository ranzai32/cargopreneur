import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Badge } from '../components/Badge'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageLayout } from '../components/PageLayout'
import { ReviewForm, ReviewList, formatRating } from '../components/Reviews'
import { api } from '../lib/api'
import { useAppState } from '../lib/app-state'
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
    Promise.all([
      fetch(`${apiBase()}/suppliers/${supplierId}`).then(readJson<ApiSupplier>),
      api.supplierReviews(supplierId),
      loadProducts({ supplierId, pageSize: 50, sort: 'rating' }),
    ]).then(([found, reviewItems, catalog]) => {
      if (cancel) return
      setSupplier(found)
      setReviews(reviewItems)
      setProducts(catalog.items)
      document.title = `${found.name} · CargoPrener`
    }).catch(() => {
      if (!cancel) setError('Supplier not found')
    })
    return () => {
      cancel = true
    }
  }, [loadProducts, supplierId])

  async function reloadReviews() {
    setReviews(await api.supplierReviews(supplierId))
    const found = await fetch(`${apiBase()}/suppliers/${supplierId}`).then(readJson<ApiSupplier>)
    setSupplier(found)
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
              <p className="mt-2 text-sm text-ink-700">{supplier.region} · {formatRating(supplier.rating, supplier.reviewCount)}</p>
              {supplier.description && <p className="mt-4 max-w-2xl text-ink-800">{supplier.description}</p>}
              <h2 className="mt-8 font-serif text-2xl text-ink-950">Catalog</h2>
              <ul className="mt-4 grid gap-4 sm:grid-cols-2">
                {products.map((product) => (
                  <li key={product.id}>
                    <Link to={`/catalog/${product.id}`} className="block h-full">
                      <Card className="h-full p-4">
                        {product.imageUrl && <img src={mediaUrl(product.imageUrl)} alt="" className="mb-3 aspect-[4/3] w-full rounded-md object-cover" />}
                        <h3 className="font-serif text-xl text-ink-950">{product.name}</h3>
                        <p className="mt-1 text-sm text-ink-700">{formatRating(product.rating, product.reviewCount)}</p>
                      </Card>
                    </Link>
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

function apiBase() {
  return import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
}

async function readJson<T>(response: Response) {
  if (!response.ok) throw new Error('request failed')
  return response.json() as Promise<T>
}
