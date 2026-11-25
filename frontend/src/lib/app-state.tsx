import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { exampleCategories, exampleProducts, exampleSuppliers, initialExampleOrders } from '../data/exampleCatalog'
import type { ApiOrder, ApiProduct, ApiSupplier, ApiUser, FieldErrors, ProductFilters, ProductPage } from '../types/catalog'
import { ApiError, api, exampleMode } from './api'
import { filterProducts } from './catalog'

const userKey = 'cargoprener-user'
const orderKey = 'cargoprener-orders'

type RegisterInput = {
  businessName: string
  contactName: string
  email: string
  password: string
  region: string
}

type SupplierInput = {
  name: string
  description: string
  region: string
}

type ProductInput = {
  supplierId: string
  name: string
  description: string
  category: string
  tags: string[]
  unit: string
  retailPrice: number
  priceTiers: Array<{ minQty: number; unitPrice: number }>
}

type AppState = {
  ready: boolean
  exampleMode: boolean
  user: ApiUser | null
  login: (email: string, password: string) => Promise<void>
  register: (input: RegisterInput) => Promise<void>
  logout: () => Promise<void>
  loadProducts: (filters: ProductFilters) => Promise<ProductPage>
  loadProduct: (id: string) => Promise<ApiProduct>
  loadCategories: () => Promise<Array<{ name: string; count: number }>>
  loadOrders: (status?: string) => Promise<ApiOrder[]>
  loadOrder: (id: string) => Promise<ApiOrder>
  createOrder: (input: { productId: string; moq: number; deadline: string; quantity: number }) => Promise<ApiOrder>
  joinOrder: (id: string, quantity: number) => Promise<ApiOrder>
  pay: (id: string, shippingAddress: string) => Promise<ApiOrder>
  loadAdminSuppliers: () => Promise<ApiSupplier[]>
  createSupplier: (input: SupplierInput) => Promise<ApiSupplier>
  verifySupplier: (id: string) => Promise<ApiSupplier>
  createProduct: (input: ProductInput) => Promise<ApiProduct>
  loadAdminOrders: () => Promise<ApiOrder[]>
  cancelOrder: (id: string) => Promise<ApiOrder>
}

const AppStateContext = createContext<AppState | null>(null)

export function AppStateProvider({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(false)
  const [user, setUser] = useState<ApiUser | null>(null)
  const [orders, setOrders] = useState<ApiOrder[]>(() => {
    if (!exampleMode) return initialExampleOrders()
    return readJson<ApiOrder[]>(orderKey) ?? initialExampleOrders()
  })
  const [suppliers, setSuppliers] = useState<ApiSupplier[]>(() => [
    ...exampleSuppliers,
    {
      id: '66666666-6666-4666-8666-666666666666',
      name: 'QuickLot Imports',
      description: 'Awaiting review.',
      region: 'Newark, NJ',
      isVerified: false,
      productCount: 0,
    },
  ])

  useEffect(() => {
    if (exampleMode) {
      setUser(readJson<ApiUser>(userKey))
      const savedOrders = readJson<ApiOrder[]>(orderKey)
      if (savedOrders) setOrders(savedOrders)
      setReady(true)
      return
    }
    api.me().then(setUser).catch(() => setUser(null)).finally(() => setReady(true))
  }, [])

  useEffect(() => {
    if (!exampleMode || !ready) return
    sessionStorage.setItem(orderKey, JSON.stringify(orders))
  }, [orders, ready])

  const value = useMemo<AppState>(() => ({
    ready,
    exampleMode,
    user,
    async login(email, password) {
      if (exampleMode) {
        const next = mockUser(email)
        sessionStorage.setItem(userKey, JSON.stringify(next))
        setUser(next)
        return
      }
      setUser(await api.login({ email, password }))
    },
    async register(input) {
      if (exampleMode) {
        const next = {
          ...mockUser(input.email),
          businessName: input.businessName,
          contactName: input.contactName,
          region: input.region,
        }
        sessionStorage.setItem(userKey, JSON.stringify(next))
        setUser(next)
        return
      }
      setUser(await api.register(input))
    },
    async logout() {
      if (!exampleMode) await api.logout().catch(() => undefined)
      sessionStorage.removeItem(userKey)
      setUser(null)
    },
    async loadProducts(filters) {
      if (exampleMode) return filterProducts(exampleProducts, filters)
      return api.products(filters)
    },
    async loadProduct(id) {
      if (exampleMode) {
        const found = exampleProducts.find((item) => item.id === id)
        if (!found) throw new ApiError(404, 'Product not found')
        return found
      }
      return api.product(id)
    },
    async loadCategories() {
      if (exampleMode) return exampleCategories()
      return api.categories()
    },
    async loadOrders(status) {
      const source = exampleMode ? orders : await api.orders(status)
      if (!status || status === 'all') return source
      return source.filter((item) => item.status === status)
    },
    async loadOrder(id) {
      if (exampleMode) {
        const found = orders.find((item) => item.id === id)
        if (!found) throw new ApiError(404, 'Order not found')
        return found
      }
      return api.order(id)
    },
    async createOrder(input) {
      if (!user) throw new ApiError(401, 'Authentication required')
      if (!exampleMode) return api.createOrder(input)
      const product = exampleProducts.find((item) => item.id === input.productId)
      if (!product) throw new ApiError(400, 'Choose a product', { productId: 'Choose a product from the catalog' })
      const created: ApiOrder = {
        id: crypto.randomUUID(),
        status: input.quantity >= input.moq ? 'confirmed' : 'open',
        moq: input.moq,
        deadline: new Date(`${input.deadline}T23:59:59`).toISOString(),
        createdAt: new Date().toISOString(),
        collectedQty: input.quantity,
        participantCount: 1,
        yourQuantity: input.quantity,
        yourPaymentStatus: 'unpaid',
        allPaid: false,
        product,
      }
      setOrders((current) => [created, ...current])
      return created
    },
    async joinOrder(id, quantity) {
      if (!user) throw new ApiError(401, 'Authentication required')
      if (!exampleMode) return api.joinOrder(id, quantity)
      const current = orders.find((item) => item.id === id)
      if (!current) throw new ApiError(404, 'Order not found')
      if (current.status !== 'open') throw new ApiError(400, 'This order is no longer open')
      if (current.yourQuantity) {
        throw new ApiError(409, 'You already joined this order', { quantity: 'You already joined this order' })
      }
      const collectedQty = current.collectedQty + quantity
      const updated: ApiOrder = {
        ...current,
        collectedQty,
        participantCount: current.participantCount + 1,
        yourQuantity: quantity,
        yourPaymentStatus: 'unpaid',
        allPaid: false,
        status: collectedQty >= current.moq ? 'confirmed' : 'open',
      }
      setOrders((items) => items.map((item) => (item.id === id ? updated : item)))
      return updated
    },
    async pay(id, shippingAddress) {
      if (!user) throw new ApiError(401, 'Authentication required')
      const address = shippingAddress.trim()
      if (!address) {
        throw new ApiError(422, 'Check the highlighted fields', { shippingAddress: 'Enter a shipping address' })
      }
      if (!exampleMode) return api.pay(id, address)
      const current = orders.find((item) => item.id === id)
      if (!current) throw new ApiError(404, 'Order not found')
      if (current.status !== 'confirmed') throw new ApiError(400, 'This order is not ready for payment')
      if (!current.yourQuantity) throw new ApiError(403, 'You are not part of this order')
      if (current.yourPaymentStatus === 'paid') throw new ApiError(409, 'This order is already paid')
      const allPaid = current.participantCount <= 1
      const updated: ApiOrder = {
        ...current,
        yourPaymentStatus: 'paid',
        allPaid,
        status: allPaid ? 'paid' : 'confirmed',
      }
      setOrders((items) => items.map((item) => (item.id === id ? updated : item)))
      return updated
    },
    async loadAdminSuppliers() {
      if (user?.role !== 'admin') throw new ApiError(403, 'You do not have access to this')
      if (!exampleMode) return api.adminSuppliers()
      return [...suppliers].sort((a, b) => a.name.localeCompare(b.name))
    },
    async createSupplier(input) {
      if (user?.role !== 'admin') throw new ApiError(403, 'You do not have access to this')
      if (!exampleMode) return api.createSupplier(input)
      const created: ApiSupplier = {
        id: crypto.randomUUID(),
        name: input.name.trim(),
        description: input.description.trim(),
        region: input.region.trim(),
        isVerified: false,
        productCount: 0,
      }
      setSuppliers((current) => [...current, created])
      return created
    },
    async verifySupplier(id) {
      if (user?.role !== 'admin') throw new ApiError(403, 'You do not have access to this')
      if (!exampleMode) return api.verifySupplier(id)
      const found = suppliers.find((item) => item.id === id)
      if (!found) throw new ApiError(404, 'Supplier not found')
      const updated = { ...found, isVerified: true }
      setSuppliers((current) => current.map((item) => (item.id === id ? updated : item)))
      return updated
    },
    async createProduct(input) {
      if (user?.role !== 'admin') throw new ApiError(403, 'You do not have access to this')
      if (!exampleMode) return api.createProduct(input)
      const supplier = suppliers.find((item) => item.id === input.supplierId)
      if (!supplier) throw new ApiError(422, 'Check the highlighted fields', { supplierId: 'Choose a supplier' })
      setSuppliers((current) => current.map((item) => (
        item.id === supplier.id ? { ...item, productCount: (item.productCount ?? 0) + 1 } : item
      )))
      return {
        id: crypto.randomUUID(),
        name: input.name.trim(),
        description: input.description.trim(),
        category: input.category.trim(),
        tags: input.tags,
        unit: input.unit.trim(),
        retailPrice: input.retailPrice,
        supplier: {
          id: supplier.id,
          name: supplier.name,
          region: supplier.region,
          isVerified: supplier.isVerified,
        },
        priceTiers: [...input.priceTiers].sort((left, right) => left.minQty - right.minQty),
        createdAt: new Date().toISOString(),
      }
    },
    async loadAdminOrders() {
      if (user?.role !== 'admin') throw new ApiError(403, 'You do not have access to this')
      if (!exampleMode) return api.adminOrders()
      return orders
    },
    async cancelOrder(id) {
      if (user?.role !== 'admin') throw new ApiError(403, 'You do not have access to this')
      if (!exampleMode) return api.cancelOrder(id)
      const current = orders.find((item) => item.id === id)
      if (!current) throw new ApiError(404, 'Order not found')
      if (current.status !== 'open' && current.status !== 'confirmed') {
        throw new ApiError(400, 'This order can no longer be cancelled')
      }
      const updated: ApiOrder = { ...current, status: 'cancelled' }
      setOrders((items) => items.map((item) => (item.id === id ? updated : item)))
      return updated
    },
  }), [orders, ready, suppliers, user])

  return <AppStateContext.Provider value={value}>{children}</AppStateContext.Provider>
}

export function useAppState() {
  const value = useContext(AppStateContext)
  if (!value) throw new Error('App state is missing')
  return value
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}

export function fieldErrorsOf(error: unknown): FieldErrors {
  return isApiError(error) && error.fieldErrors ? error.fieldErrors : {}
}

function mockUser(email: string): ApiUser {
  return {
    id: `local-${email.toLowerCase()}`,
    businessName: 'Your business',
    contactName: 'You',
    email: email.toLowerCase(),
    region: 'Local example',
    role: 'business',
    createdAt: new Date().toISOString(),
  }
}

function readJson<T>(key: string): T | null {
  const raw = sessionStorage.getItem(key)
  if (!raw) return null
  try {
    return JSON.parse(raw) as T
  } catch {
    return null
  }
}
