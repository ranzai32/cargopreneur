import type { ApiOrder, ApiProduct, ApiSupplier } from '../types/catalog'

const harbor: ApiProduct['supplier'] = {
  id: '11111111-1111-4111-8111-111111111111',
  name: 'Harbor Mill Supply',
  region: 'Grand Rapids, MI',
  isVerified: true,
}

const northline: ApiProduct['supplier'] = {
  id: '22222222-2222-4222-8222-222222222222',
  name: 'Northline Paper',
  region: 'Columbus, OH',
  isVerified: true,
}

const lakeshore: ApiProduct['supplier'] = {
  id: '55555555-5555-4555-8555-555555555555',
  name: 'Lakeshore Janitorial',
  region: 'Milwaukee, WI',
  isVerified: true,
}

export const exampleSuppliers: ApiSupplier[] = [
  { ...harbor, description: 'Packaging supplier reviewed for the catalog.', productCount: 3 },
  { ...northline, description: 'Paper goods supplier reviewed for the catalog.', productCount: 2 },
  { ...lakeshore, description: 'Cleaning supplies reviewed for shops and workshops.', productCount: 3 },
]

export const exampleProducts: ApiProduct[] = [
  product(harbor, '33333333-3333-4333-8333-333333333333', 'Kraft mailer boxes, 9 × 6 × 3 in', 'Carton of 50 mailer boxes.', 'Packaging', ['mailer', 'boxes'], 'carton', 16.5, [[100, 11.2], [140, 9.6], [180, 8.4]], '2026-07-01T15:00:00Z'),
  product(northline, '44444444-4444-4444-8444-444444444444', 'Shop towels, blue, case of 50', 'Case of 50 blue shop towels.', 'Workshop supplies', ['towels', 'workshop'], 'case', 28, [[40, 19.5], [80, 16.75], [120, 14.25]], '2026-07-15T15:00:00Z'),
  product(harbor, '77777777-7777-4777-8777-777777777701', 'Stretch wrap film, 18 in x 1500 ft', 'Clear pallet wrap sold by the roll.', 'Packaging', ['wrap', 'pallet'], 'roll', 12.4, [[50, 8.1], [100, 6.9], [200, 5.75]], '2026-08-01T15:00:00Z'),
  product(harbor, '77777777-7777-4777-8777-777777777702', 'Corrugated boxes, 12 x 12 x 12 in', 'Bundle of 25 shipping boxes.', 'Packaging', ['boxes', 'shipping'], 'bundle', 22, [[40, 15.4], [80, 13.1], [160, 11.25]], '2026-08-10T15:00:00Z'),
  product(northline, '77777777-7777-4777-8777-777777777703', 'Letter copy paper, 10-ream case', 'Case of ten reams of letter paper.', 'Store supplies', ['paper', 'office'], 'case', 42, [[20, 31.5], [40, 28], [80, 24.75]], '2026-08-18T15:00:00Z'),
  product(lakeshore, '77777777-7777-4777-8777-777777777704', 'Nitrile gloves, medium, box of 100', 'Powder-free gloves for workshop and counter use.', 'Workshop supplies', ['gloves', 'safety'], 'box', 14.8, [[24, 9.9], [48, 8.4], [96, 7.15]], '2026-08-25T15:00:00Z'),
  product(lakeshore, '77777777-7777-4777-8777-777777777705', 'Glass cleaner concentrate, 1 gallon', 'Dilute for counters and front windows.', 'Cleaning', ['cleaner', 'glass'], 'gallon', 18.25, [[12, 12.4], [24, 10.8], [48, 9.35]], '2026-09-02T15:00:00Z'),
  product(lakeshore, '77777777-7777-4777-8777-777777777706', 'Microfiber cloths, pack of 24', 'Reusable cloths for glass and fixtures.', 'Cleaning', ['cloths', 'cleaning'], 'pack', 16, [[20, 11.2], [40, 9.5], [80, 8.1]], '2026-09-08T15:00:00Z'),
]

export function exampleCategories() {
  const counts = new Map<string, number>()
  for (const item of exampleProducts) {
    counts.set(item.category, (counts.get(item.category) ?? 0) + 1)
  }
  return [...counts.entries()]
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => a.name.localeCompare(b.name))
}

export function initialExampleOrders(): ApiOrder[] {
  return [
    order('a1a1a1a1-a1a1-4a1a-8a1a-a1a1a1a1a1a1', '33333333-3333-4333-8333-333333333333', 'open', 100, 78, 6, 14),
    order('b2b2b2b2-b2b2-4b2b-8b2b-b2b2b2b2b2b2', '44444444-4444-4444-8444-444444444444', 'open', 40, 18, 3, 10),
    order('c3c3c3c3-c3c3-4c3c-8c3c-c3c3c3c3c3c3', '77777777-7777-4777-8777-777777777704', 'open', 48, 24, 2, 21),
    order('d4d4d4d4-d4d4-4d4d-8d4d-d4d4d4d4d4d4', '77777777-7777-4777-8777-777777777701', 'confirmed', 50, 50, 4, 6),
  ]
}

function product(
  supplier: ApiProduct['supplier'],
  id: string,
  name: string,
  description: string,
  category: string,
  tags: string[],
  unit: string,
  retailPrice: number,
  tiers: Array<[number, number]>,
  createdAt: string,
): ApiProduct {
  return {
    id,
    name,
    description,
    category,
    tags,
    unit,
    retailPrice,
    supplier,
    priceTiers: tiers.map(([minQty, unitPrice]) => ({ minQty, unitPrice })),
    createdAt,
  }
}

function order(
  id: string,
  productId: string,
  status: ApiOrder['status'],
  moq: number,
  collectedQty: number,
  participantCount: number,
  daysAhead: number,
): ApiOrder {
  const item = exampleProducts.find((entry) => entry.id === productId)
  if (!item) throw new Error(`Missing example product ${productId}`)
  const deadline = new Date()
  deadline.setDate(deadline.getDate() + daysAhead)
  return {
    id,
    status,
    moq,
    deadline: deadline.toISOString(),
    createdAt: item.createdAt,
    collectedQty,
    participantCount,
    yourQuantity: null,
    product: item,
  }
}
