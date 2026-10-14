import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Badge } from '../components/Badge'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageHeader } from '../components/PageHeader'
import { PageLayout } from '../components/PageLayout'
import { formatMoney } from '../lib/format'
import { openingTier } from '../lib/catalog'
import { isApiError, useAppState } from '../lib/app-state'
import type { ProductPage } from '../types/catalog'

const sorts = [
  { value: 'newest', label: 'Newest' },
  { value: 'name', label: 'Name' },
  { value: 'price_asc', label: 'Lowest tier price' },
  { value: 'price_desc', label: 'Highest tier price' },
]

export function CatalogPage() {
  const { loadProducts, loadCategories } = useAppState()
  const [params, setParams] = useSearchParams()
  const [result, setResult] = useState<ProductPage | null>(null)
  const [categories, setCategories] = useState<Array<{ name: string; count: number }>>([])
  const [error, setError] = useState('')
  const q = params.get('q') ?? ''
  const category = params.get('category') ?? ''
  const sort = (params.get('sort') ?? 'newest') as 'newest' | 'name' | 'price_asc' | 'price_desc'
  const page = Number(params.get('page') ?? '1')

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
    loadProducts({ q, category, sort, page, pageSize: 12 })
      .then((pageResult) => {
        if (!cancel) setResult(pageResult)
      })
      .catch((reason) => {
        if (!cancel) setError(isApiError(reason) ? reason.message : 'The catalog could not be loaded')
      })
    return () => {
      cancel = true
    }
  }, [q, category, sort, page, loadProducts])

  function update(next: Record<string, string>) {
    const following = new URLSearchParams(params)
    for (const [key, value] of Object.entries(next)) {
      if (value) following.set(key, value)
      else following.delete(key)
    }
    if (!('page' in next)) following.delete('page')
    setParams(following)
  }

  const pages = result ? Math.max(1, Math.ceil(result.total / result.pageSize)) : 1

  return (
    <PageLayout>
      <Container className="py-section">
        <PageHeader
          title="Catalog"
          lede="Reviewed suppliers and the wholesale tiers for each product. The price drops when a pooled order reaches a higher tier."
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
        <ul className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {result?.items.map((product) => {
            const tier = openingTier(product)
            return (
              <li key={product.id}>
                <Link to={`/catalog/${product.id}`} className="block h-full rounded-xl">
                  <Card className="flex h-full flex-col p-5">
                    <Badge tone="accent">{product.category}</Badge>
                    <h2 className="mt-3 font-serif text-2xl leading-tight text-ink-950">{product.name}</h2>
                    <p className="mt-2 text-sm text-ink-700">
                      {product.supplier.name} · {product.supplier.region}
                    </p>
                    <p className="mt-4 font-semibold tabular-nums text-ink-950">
                      {tier ? `${formatMoney(tier.unitPrice)} / ${product.unit}` : 'No wholesale tier'}
                    </p>
                    <p className="text-sm text-ink-700">
                      {tier ? `At ${tier.minQty}+ · Retail ${formatMoney(product.retailPrice)}` : `Retail ${formatMoney(product.retailPrice)}`}
                    </p>
                  </Card>
                </Link>
              </li>
            )
          })}
        </ul>
        {result && result.total > result.pageSize && (
          <div className="mt-8 flex items-center justify-between gap-4">
            <button type="button" className="text-sm font-semibold text-accent-800 disabled:opacity-40" disabled={page <= 1} onClick={() => update({ page: String(page - 1) })}>
              Previous
            </button>
            <p className="text-sm text-ink-700">Page {page} of {pages}</p>
            <button type="button" className="text-sm font-semibold text-accent-800 disabled:opacity-40" disabled={page >= pages} onClick={() => update({ page: String(page + 1) })}>
              Next
            </button>
          </div>
        )}
      </Container>
    </PageLayout>
  )
}
