import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageHeader } from '../components/PageHeader'
import { PageLayout } from '../components/PageLayout'
import { formatRating } from '../components/Reviews'
import { formatMoney } from '../lib/format'
import { mediaUrl } from '../lib/media'
import { openingTier } from '../lib/catalog'
import { isApiError, useAppState } from '../lib/app-state'
import type { ApiProduct, ProductPage } from '../types/catalog'

const sorts = [
  { value: 'newest', label: 'Newest' },
  { value: 'name', label: 'Name' },
  { value: 'price_asc', label: 'Lowest tier price' },
  { value: 'price_desc', label: 'Highest tier price' },
  { value: 'rating', label: 'Reviews' },
]

export function CatalogPage() {
  const { loadProducts, loadCategories } = useAppState()
  const [params, setParams] = useSearchParams()
  const [result, setResult] = useState<ProductPage | null>(null)
  const [categories, setCategories] = useState<Array<{ name: string; count: number }>>([])
  const [error, setError] = useState('')
  const q = params.get('q') ?? ''
  const category = params.get('category') ?? ''
  const sort = (params.get('sort') ?? 'rating') as 'newest' | 'name' | 'price_asc' | 'price_desc' | 'rating'

  useEffect(() => {
    document.title = 'Catalog · CargoPrener'
  }, [])

  useEffect(() => {
    let cancel = false
    loadCategories().then((items) => {
      if (!cancel) setCategories(items)
    }).catch(() => {
      if (!cancel) setCategories([])
    })
    return () => {
      cancel = true
    }
  }, [loadCategories])

  useEffect(() => {
    let cancel = false
    setError('')
    loadProducts({ q, category, sort, page: 1, pageSize: 50 })
      .then((pageResult) => {
        if (!cancel) setResult(pageResult)
      })
      .catch((reason) => {
        if (!cancel) setError(isApiError(reason) ? reason.message : 'The catalog could not be loaded')
      })
    return () => {
      cancel = true
    }
  }, [q, category, sort, loadProducts])

  function update(next: Record<string, string>) {
    const following = new URLSearchParams(params)
    for (const [key, value] of Object.entries(next)) {
      if (value) following.set(key, value)
      else following.delete(key)
    }
    if (!('page' in next)) following.delete('page')
    setParams(following)
  }

  return (
    <PageLayout>
      <Container className="py-section">
        <PageHeader
          title="Catalog"
          lede="Each category mixes goods from reviewed suppliers. Sort a category by reviews, price, or name."
        />
        <form className="mt-8 grid gap-3 md:grid-cols-[1.4fr_1fr_1fr]" onSubmit={(event) => event.preventDefault()}>
          <label className="block">
            <span className="text-sm font-semibold text-ink-900">Search</span>
            <input
              value={q}
              onChange={(event) => update({ q: event.target.value })}
              className="mt-1.5 h-11 w-full rounded-md border border-ink-300 bg-white px-3"
              placeholder="Boxes, gloves, paper"
            />
          </label>
          <label className="block">
            <span className="text-sm font-semibold text-ink-900">Category</span>
            <select
              value={category}
              onChange={(event) => update({ category: event.target.value })}
              className="mt-1.5 h-11 w-full rounded-md border border-ink-300 bg-white px-3"
            >
              <option value="">All categories</option>
              {categories.map((item) => (
                <option key={item.name} value={item.name}>
                  {item.name} ({item.count})
                </option>
              ))}
            </select>
          </label>
          <label className="block">
            <span className="text-sm font-semibold text-ink-900">Sort</span>
            <select
              value={sort}
              onChange={(event) => update({ sort: event.target.value })}
              className="mt-1.5 h-11 w-full rounded-md border border-ink-300 bg-white px-3"
            >
              {sorts.map((item) => (
                <option key={item.value} value={item.value}>{item.label}</option>
              ))}
            </select>
          </label>
        </form>
        {error && <p role="alert" className="mt-6 text-danger">{error}</p>}
        {!result && !error && <p className="mt-8 text-ink-700">Loading the catalog…</p>}
        {result && result.items.length === 0 && (
          <p className="mt-8 text-ink-700">No products match that search.</p>
        )}
        <div className="mt-10 grid gap-12">
          {groups(result?.items ?? [], sort).map((group) => (
            <section key={group.name}>
              <h2 className="font-serif text-3xl text-ink-950">{group.name}</h2>
              <ul className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {group.items.map((product) => {
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
                          <h3 className="font-serif text-2xl leading-tight text-ink-950">{product.name}</h3>
                        </Link>
                        <p className="mt-2 text-sm text-ink-700">{formatRating(product.rating, product.reviewCount)}</p>
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
          ))}
        </div>
      </Container>
    </PageLayout>
  )
}

function groups(items: ApiProduct[], sort: string) {
  const buckets = new Map<string, ApiProduct[]>()
  for (const product of items) {
    const list = buckets.get(product.category) ?? []
    list.push(product)
    buckets.set(product.category, list)
  }
  return [...buckets.entries()]
    .sort((left, right) => left[0].localeCompare(right[0]))
    .map(([name, group]) => ({
      name,
      items: [...group].sort((left, right) => compareProducts(left, right, sort)),
    }))
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
