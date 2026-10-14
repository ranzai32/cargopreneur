import type { ApiProduct, ProductFilters, ProductPage } from '../types/catalog'

export function filterProducts(products: ApiProduct[], filters: ProductFilters): ProductPage {
  const page = filters.page && filters.page > 0 ? filters.page : 1
  const pageSize = filters.pageSize && filters.pageSize > 0 ? Math.min(filters.pageSize, 50) : 12
  const query = filters.q?.trim().toLowerCase() ?? ''
  let items = products.filter((product) => product.supplier.isVerified)
  if (query) {
    items = items.filter((product) => {
      const haystack = [product.name, product.description, ...product.tags].join(' ').toLowerCase()
      return haystack.includes(query)
    })
  }
  if (filters.category) {
    items = items.filter((product) => product.category.toLowerCase() === filters.category?.toLowerCase())
  }
  items = [...items].sort((left, right) => compareProducts(left, right, filters.sort ?? 'newest'))
  const total = items.length
  const start = (page - 1) * pageSize
  return { items: items.slice(start, start + pageSize), total, page, pageSize }
}

function compareProducts(left: ApiProduct, right: ApiProduct, sort: NonNullable<ProductFilters['sort']>) {
  if (sort === 'name') return left.name.localeCompare(right.name)
  if (sort === 'price_asc' || sort === 'price_desc') {
    const delta = lowestPrice(left) - lowestPrice(right)
    return sort === 'price_asc' ? delta : -delta
  }
  return right.createdAt.localeCompare(left.createdAt)
}

function lowestPrice(product: ApiProduct) {
  return product.priceTiers.reduce((lowest, tier) => Math.min(lowest, tier.unitPrice), Number.POSITIVE_INFINITY)
}

export function openingTier(product: ApiProduct) {
  return [...product.priceTiers].sort((a, b) => a.minQty - b.minQty)[0]
}
