import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageLayout } from '../components/PageLayout'
import { RatingLink } from '../components/Reviews'
import { formatMoney } from '../lib/format'
import { mediaUrl } from '../lib/media'
import { openingTier } from '../lib/catalog'
import { isApiError, useAppState } from '../lib/app-state'
import type { ApiProduct } from '../types/catalog'

const sorts = [
  { value: 'rating', label: 'Reviews' },
  { value: 'newest', label: 'Newest' },
  { value: 'name', label: 'Name' },
  { value: 'price_asc', label: 'Lowest tier price' },
  { value: 'price_desc', label: 'Highest tier price' },
]

export function CatalogPage() {
  const { loadProducts } = useAppState()
  const [params, setParams] = useSearchParams()
  const [products, setProducts] = useState<ApiProduct[] | null>(null)
  const [error, setError] = useState('')
  const q = params.get('q') ?? ''
  const category = params.get('category') ?? ''
  const sort = params.get('sort') ?? 'rating'
  const browsing = !q && !category

  useEffect(() => {
    document.title = category ? `${category} · CargoPrener` : 'Categories · CargoPrener'
  }, [category])

  useEffect(() => {
    let cancel = false
    setError('')
    loadProducts({ page: 1, pageSize: 50, sort: 'rating' })
      .then((pageResult) => {
        if (!cancel) setProducts(pageResult.items)
      })
      .catch((reason) => {
        if (!cancel) setError(isApiError(reason) ? reason.message : 'The catalog could not be loaded')
      })
    return () => {
      cancel = true
    }
  }, [loadProducts])

  function update(next: Record<string, string>) {
    const following = new URLSearchParams(params)
    for (const [key, value] of Object.entries(next)) {
      if (value) following.set(key, value)
      else following.delete(key)
    }
    setParams(following)
  }

  const visible = (products ?? []).filter((product) => {
    if (category && product.category.toLowerCase() !== category.toLowerCase()) return false
    if (!q.trim()) return true
    const haystack = [product.name, product.description, product.category, product.supplier.name, ...product.tags].join(' ').toLowerCase()
    return haystack.includes(q.trim().toLowerCase())
  })
  const tiles = categoryTiles(products ?? [])

  return (
    <PageLayout>
      <Container className="py-8 sm:py-section">
        <label className="block">
          <span className="sr-only">Search products</span>
          <input
            value={q}
            onChange={(event) => update({ q: event.target.value, category: '' })}
            placeholder="Search products"
            className="h-12 w-full rounded-full border border-ink-200 bg-white px-5 text-base text-ink-950 shadow-sm"
          />
        </label>
        {error && <p role="alert" className="mt-6 text-danger">{error}</p>}
        {!products && !error && <p className="mt-8 text-ink-700">Loading the catalog…</p>}
        {products && browsing && (
          <section className="mt-6">
            <h1 className="font-serif text-3xl text-ink-950">Categories</h1>
            <ul className="mt-4 grid grid-cols-3 gap-2.5 sm:gap-3">
              {tiles.map((tile) => (
                <li key={tile.name}>
                  <button
                    type="button"
                    onClick={() => update({ category: tile.name, q: '' })}
                    className="flex h-36 w-full flex-col rounded-2xl bg-[#eceae6] p-2.5 text-left sm:h-44"
                  >
                    <span className="text-[13px] font-semibold leading-tight text-ink-950 sm:text-sm">{tile.name}</span>
                    {tile.imageUrl ? (
                      <span className="mt-1 flex min-h-0 flex-1 items-end justify-center">
                        <img src={mediaUrl(tile.imageUrl)} alt="" className="max-h-full max-w-full object-contain" />
                      </span>
                    ) : (
                      <span className="mt-auto text-xs font-semibold text-ink-500">{tile.count}</span>
                    )}
                  </button>
                </li>
              ))}
            </ul>
          </section>
        )}
        {products && !browsing && (
          <section className="mt-6">
            <button type="button" onClick={() => update({ category: '', q: '' })} className="text-sm font-semibold text-accent-800">
              All categories
            </button>
            <div className="mt-3 flex flex-wrap items-end justify-between gap-3">
              <h1 className="font-serif text-3xl text-ink-950">{category || 'Search'}</h1>
              <label className="text-sm font-semibold text-ink-900">
                <span className="sr-only">Sort</span>
                <select
                  value={sort}
                  onChange={(event) => update({ sort: event.target.value })}
                  className="h-11 rounded-full border border-ink-200 bg-white px-4"
                >
                  {sorts.map((item) => (
                    <option key={item.value} value={item.value}>{item.label}</option>
                  ))}
                </select>
              </label>
            </div>
            {visible.length === 0 && <p className="mt-8 text-ink-700">No products match that search.</p>}
            <ul className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {[...visible].sort((left, right) => compareProducts(left, right, sort)).map((product) => {
                const tier = openingTier(product)
                return (
                  <li key={product.id}>
                    <Card className="flex h-full flex-col p-5">
                      <Link to={`/catalog/${product.id}`}>
                        {product.imageUrl ? (
                          <img src={mediaUrl(product.imageUrl)} alt="" className="mb-4 aspect-[4/3] w-full rounded-md object-cover" />
                        ) : (
                          <div className="mb-4 flex aspect-[4/3] items-center justify-center rounded-md bg-ink-100 text-sm font-semibold text-ink-600">{product.category}</div>
                        )}
                        <h2 className="font-serif text-2xl leading-tight text-ink-950">{product.name}</h2>
                      </Link>
                      <p className="mt-2">
                        <RatingLink to={`/catalog/${product.id}/reviews`} rating={product.rating} count={product.reviewCount} />
                      </p>
                      <p className="mt-2 text-sm text-ink-700">
                        <Link to={`/suppliers/${product.supplier.id}`} className="font-semibold text-accent-800">{product.supplier.name}</Link>
                        {' · '}{product.supplier.region}
                      </p>
                      <p className="mt-4 font-semibold tabular-nums text-ink-950">
                        {tier ? `${formatMoney(tier.unitPrice)} / ${product.unit}` : 'No wholesale tier'}
                      </p>
                    </Card>
                  </li>
                )
              })}
            </ul>
          </section>
        )}
      </Container>
    </PageLayout>
  )
}

function categoryTiles(products: ApiProduct[]) {
  const tiles = new Map<string, { name: string; count: number; imageUrl: string | null; reviews: number }>()
  for (const product of products) {
    const current = tiles.get(product.category)
    const betterImage = !current?.imageUrl || product.reviewCount > current.reviews
    tiles.set(product.category, {
      name: product.category,
      count: (current?.count ?? 0) + 1,
      imageUrl: betterImage && product.imageUrl ? product.imageUrl : current?.imageUrl ?? product.imageUrl,
      reviews: betterImage ? product.reviewCount : current?.reviews ?? 0,
    })
  }
  return [...tiles.values()].sort((left, right) => left.name.localeCompare(right.name))
}

function compareProducts(left: ApiProduct, right: ApiProduct, sort: string) {
  if (sort === 'name') return left.name.localeCompare(right.name)
  if (sort === 'price_asc' || sort === 'price_desc') {
    const price = (product: ApiProduct) => product.priceTiers.reduce((lowest, tier) => Math.min(lowest, tier.unitPrice), Number.POSITIVE_INFINITY)
    const delta = price(left) - price(right)
    return sort === 'price_asc' ? delta : -delta
  }
  if (sort === 'rating') return right.rating - left.rating || right.reviewCount - left.reviewCount
  return right.createdAt.localeCompare(left.createdAt)
}
