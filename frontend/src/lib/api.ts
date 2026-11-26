import type { AdminDispute, ApiOrder, ApiProduct, ApiSupplier, ApiUser, FieldErrors, ProductFilters, ProductPage } from '../types/catalog'

export const apiBase = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
export const exampleMode = import.meta.env.VITE_USE_MOCKS !== 'false'

export class ApiError extends Error {
  status: number
  fieldErrors?: FieldErrors

  constructor(status: number, message: string, fieldErrors?: FieldErrors) {
    super(message)
    this.status = status
    this.fieldErrors = fieldErrors
  }
}

export const api = {
  me: () => request<ApiUser>('/auth/me'),
  login: (body: { email: string; password: string }) =>
    request<ApiUser>('/auth/login', { method: 'POST', body }),
  register: (body: {
    businessName: string
    contactName: string
    email: string
    password: string
    region: string
  }) => request<ApiUser>('/auth/register', { method: 'POST', body }),
  logout: () => request<void>('/auth/logout', { method: 'POST' }),
  products: (filters: ProductFilters) => request<ProductPage>(`/products${queryString(filters)}`),
  product: (id: string) => request<ApiProduct>(`/products/${id}`),
  categories: () => request<Array<{ name: string; count: number }>>('/categories'),
  orders: (status?: string) => request<ApiOrder[]>(`/orders${status ? `?status=${status}` : ''}`),
  order: (id: string) => request<ApiOrder>(`/orders/${id}`),
  createOrder: (body: { productId: string; moq: number; deadline: string; quantity: number }) =>
    request<ApiOrder>('/orders', { method: 'POST', body }),
  joinOrder: (id: string, quantity: number) =>
    request<ApiOrder>(`/orders/${id}/join`, { method: 'POST', body: { quantity } }),
  pay: (id: string, shippingAddress: string) =>
    request<ApiOrder>(`/orders/${id}/pay`, { method: 'POST', body: { shippingAddress } }),
  addresses: (q: string) => request<{ items: string[] }>(`/addresses?q=${encodeURIComponent(q)}`),
  uploadPhoto: (file: File) => uploadFile(file),
  adminSuppliers: () => request<ApiSupplier[]>('/admin/suppliers'),
  createSupplier: (body: { name: string; description: string; region: string }) =>
    request<ApiSupplier>('/admin/suppliers', { method: 'POST', body }),
  verifySupplier: (id: string) => request<ApiSupplier>(`/admin/suppliers/${id}/verify`, { method: 'POST' }),
  createProduct: (body: {
    supplierId: string
    name: string
    description: string
    category: string
    tags: string[]
    unit: string
    retailPrice: number
    imageUrl?: string
    priceTiers: Array<{ minQty: number; unitPrice: number }>
  }) => request<ApiProduct>('/admin/products', { method: 'POST', body }),
  adminOrders: () => request<ApiOrder[]>('/admin/orders'),
  cancelOrder: (id: string) => request<ApiOrder>(`/admin/orders/${id}/cancel`, { method: 'POST' }),
  openDispute: (id: string, reason: string) =>
    request<ApiOrder>(`/orders/${id}/disputes`, { method: 'POST', body: { reason } }),
  adminDisputes: () => request<AdminDispute[]>('/admin/disputes'),
  resolveDispute: (id: string, outcome: string) =>
    request<AdminDispute>(`/admin/disputes/${id}/resolve`, { method: 'POST', body: { outcome } }),
  shipParticipation: (id: string) => request<ApiOrder>(`/admin/participations/${id}/ship`, { method: 'POST' }),
  deliverParticipation: (id: string) => request<ApiOrder>(`/admin/participations/${id}/deliver`, { method: 'POST' }),
}

async function uploadFile(file: File) {
  const body = new FormData()
  body.append('file', file)
  const response = await fetch(`${apiBase}/admin/uploads`, {
    method: 'POST',
    credentials: 'include',
    body,
  })
  const payload = (await response.json().catch(() => null)) as { url?: string; message?: string } | null
  if (!response.ok || !payload?.url) {
    throw new ApiError(response.status, payload?.message || 'Photo could not be uploaded')
  }
  return payload.url
}

async function request<T>(path: string, init?: { method?: string; body?: unknown }): Promise<T> {
  const response = await fetch(`${apiBase}${path}`, {
    method: init?.method ?? 'GET',
    credentials: 'include',
    headers: init?.body ? { 'Content-Type': 'application/json' } : undefined,
    body: init?.body ? JSON.stringify(init.body) : undefined,
  })
  if (response.status === 204) return undefined as T
  const payload = (await response.json().catch(() => null)) as {
    message?: string
    fieldErrors?: FieldErrors
  } | null
  if (!response.ok) {
    throw new ApiError(response.status, payload?.message || 'Request failed', payload?.fieldErrors)
  }
  return payload as T
}

function queryString(filters: ProductFilters) {
  const params = new URLSearchParams()
  if (filters.q) params.set('q', filters.q)
  if (filters.category) params.set('category', filters.category)
  if (filters.sort) params.set('sort', filters.sort)
  if (filters.page) params.set('page', String(filters.page))
  if (filters.pageSize) params.set('pageSize', String(filters.pageSize))
  const value = params.toString()
  return value ? `?${value}` : ''
}
