export type ApiSupplier = {
  id: string
  name: string
  description?: string
  region: string
  isVerified: boolean
  productCount?: number
  rating?: number
  reviewCount?: number
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
  imageUrl: string | null
  rating: number
  reviewCount: number
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
  role: 'business' | 'admin' | 'supplier'
  supplierId: string | null
  createdAt: string
}

export type DeliveryStatus = 'pending' | 'shipped' | 'delivered'

export type ApiParticipation = {
  id: string
  quantity: number
  shippingAddress: string
  paymentStatus: 'unpaid' | 'paid'
  deliveryStatus: DeliveryStatus
}

export type ApiReview = {
  id: string
  author: string
  rating: number
  body: string
  createdAt: string
}

export type ApiDispute = {
  id: string
  orderId: string
  openedBy: string
  businessName: string
  reason: string
  status: 'open' | 'resolved'
  outcome: string
  createdAt: string
  resolvedAt: string | null
}

export type AdminDispute = ApiDispute & {
  productName: string
  orderStatus: ApiOrder['status']
}

export type ApiOrder = {
  id: string
  status: 'open' | 'confirmed' | 'paid' | 'completed' | 'cancelled'
  moq: number
  deadline: string
  createdAt: string
  collectedQty: number
  participantCount: number
  yourQuantity: number | null
  yourPaymentStatus: 'unpaid' | 'paid' | null
  yourDeliveryStatus: DeliveryStatus | null
  allPaid: boolean
  participants: ApiParticipation[]
  disputes: ApiDispute[]
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
  supplierId?: string
  sort?: 'newest' | 'name' | 'price_asc' | 'price_desc' | 'rating'
  page?: number
  pageSize?: number
}

export type FieldErrors = Record<string, string>
