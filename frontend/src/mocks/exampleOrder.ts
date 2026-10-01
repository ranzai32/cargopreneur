import type { Order, Product, Supplier } from '../types'

function isoDaysFromNow(days: number): string {
  const date = new Date()
  date.setDate(date.getDate() + days)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export const exampleSupplier: Supplier = {
  id: 'sup_harbor_mill',
  name: 'Harbor Mill Supply',
  location: 'Grand Rapids, MI',
  moderated: true,
}

export const exampleProduct: Product = {
  id: 'prd_kraft_mailer',
  supplierId: exampleSupplier.id,
  name: 'Kraft mailer boxes, 9 × 6 × 3 in',
  unit: 'carton',
  retailPrice: 16.5,
  moq: 100,
  tiers: [
    { minQty: 100, price: 11.2 },
    { minQty: 140, price: 9.6 },
    { minQty: 180, price: 8.4 },
  ],
}

export const exampleOrder: Order = {
  id: 'ord_example_mailers',
  productId: exampleProduct.id,
  status: 'open',
  collectedQty: 78,
  participantCount: 6,
  deadline: isoDaysFromNow(14),
}

export const exampleQuantity = {
  min: 1,
  max: 120,
  initial: 24,
}
