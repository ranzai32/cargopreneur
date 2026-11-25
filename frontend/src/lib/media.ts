import { apiBase } from './api'

export function mediaUrl(path?: string | null) {
  if (!path) return ''
  if (path.startsWith('http://') || path.startsWith('https://')) return path
  return `${apiBase}${path}`
}
