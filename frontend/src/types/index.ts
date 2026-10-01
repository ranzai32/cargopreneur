export type OrderStatus = 'open' | 'confirmed' | 'cancelled'

export type Supplier = {
  id: string
  name: string
  location: string
  moderated: boolean
}

export type PriceTier = {
  minQty: number
  price: number
}

export type Product = {
  id: string
  supplierId: string
  name: string
  unit: string
  retailPrice: number
  moq: number
  tiers: PriceTier[]
}

export type Order = {
  id: string
  productId: string
  status: OrderStatus
  collectedQty: number
  participantCount: number
  deadline: string
}
