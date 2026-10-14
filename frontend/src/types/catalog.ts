export type ApiSupplier = {
  id: string
  name: string
  description?: string
  region: string
  isVerified: boolean
  productCount?: number
}

export type ApiPriceTier = {
  minQty: number
  unitPrice: number
}

export type ApiProduct = {
  id: string
  name: string
  description: string
  category: string
  tags: string[]
  unit: string
  retailPrice: number
  supplier: {
    id: string
    name: string
    region: string
    isVerified: boolean
  }
  priceTiers: ApiPriceTier[]
  createdAt: string
}

export type ApiUser = {
  id: string
  businessName: string
  contactName: string
  email: string
  region: string
  role: 'business' | 'admin'
  createdAt: string
}

export type ApiOrder = {
  id: string
  status: 'open' | 'confirmed' | 'cancelled'
  moq: number
  deadline: string
  createdAt: string
  collectedQty: number
  participantCount: number
  yourQuantity: number | null
  product: ApiProduct
}

export type ProductPage = {
  items: ApiProduct[]
  total: number
  page: number
  pageSize: number
}

export type ProductFilters = {
  q?: string
  category?: string
  sort?: 'newest' | 'name' | 'price_asc' | 'price_desc'
  page?: number
  pageSize?: number
}

export type FieldErrors = Record<string, string>
