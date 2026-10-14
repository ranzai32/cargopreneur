import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { exampleCategories, exampleProducts, initialExampleOrders } from '../data/exampleCatalog'
import type { ApiOrder, ApiProduct, ApiUser, FieldErrors, ProductFilters, ProductPage } from '../types/catalog'
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
}

const AppStateContext = createContext<AppState | null>(null)

export function AppStateProvider({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(false)
  const [user, setUser] = useState<ApiUser | null>(null)
  const [orders, setOrders] = useState<ApiOrder[]>(initialExampleOrders)

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
    if (!exampleMode) return
    sessionStorage.setItem(orderKey, JSON.stringify(orders))
  }, [orders])

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
      const source = exampleMode ? orders : await api.orders()
      const found = source.find((item) => item.id === id)
      if (!found) throw new ApiError(404, 'Order not found')
      return found
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
        status: collectedQty >= current.moq ? 'confirmed' : 'open',
      }
      setOrders((items) => items.map((item) => (item.id === id ? updated : item)))
      return updated
    },
  }), [orders, ready, user])

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
