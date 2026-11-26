import { ApiError, api, exampleMode } from './api'
import type { SavedAddress } from './api'

const storageKey = (userId: string) => `cargoprener-addresses:${userId}`

export async function loadAddresses(userId: string) {
  if (!exampleMode) return api.savedAddresses()
  return readBook(userId)
}

export async function createAddress(userId: string, input: { label: string; line: string }) {
  const body = cleanAddress(input)
  if (!exampleMode) return api.createAddress(body)
  const book = readBook(userId)
  if (book.length >= 12) throw new ApiError(422, 'You can save up to 12 addresses')
  const next: SavedAddress = {
    id: crypto.randomUUID(),
    label: body.label,
    line: body.line,
    updatedAt: new Date().toISOString(),
  }
  writeBook(userId, [next, ...book])
  return next
}

export async function updateAddress(userId: string, id: string, input: { label: string; line: string }) {
  const body = cleanAddress(input)
  if (!exampleMode) return api.updateAddress(id, body)
  const book = readBook(userId)
  const next = book.map((item) => (
    item.id === id ? { ...item, label: body.label, line: body.line, updatedAt: new Date().toISOString() } : item
  ))
  if (!next.some((item) => item.id === id)) throw new ApiError(404, 'Address not found')
  writeBook(userId, next)
  return next.find((item) => item.id === id) as SavedAddress
}

export async function removeAddress(userId: string, id: string) {
  if (!exampleMode) {
    await api.deleteAddress(id)
    return
  }
  writeBook(userId, readBook(userId).filter((item) => item.id !== id))
}

function cleanAddress(input: { label: string; line: string }) {
  const label = input.label.trim() || 'Address'
  const line = input.line.trim()
  const fieldErrors: Record<string, string> = {}
  if (label.length > 40) fieldErrors.label = 'Use a shorter name'
  if (line.length < 8) fieldErrors.line = 'Enter a full street address'
  if (line.length > 240) fieldErrors.line = 'Use a shorter address'
  if (Object.keys(fieldErrors).length > 0) {
    throw new ApiError(422, 'Check the highlighted fields', fieldErrors)
  }
  return { label, line }
}

function readBook(userId: string): SavedAddress[] {
  try {
    const raw = localStorage.getItem(storageKey(userId))
    const parsed = raw ? JSON.parse(raw) as SavedAddress[] : []
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function writeBook(userId: string, items: SavedAddress[]) {
  localStorage.setItem(storageKey(userId), JSON.stringify(items))
}
