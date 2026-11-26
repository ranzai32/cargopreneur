import { useEffect, useId, useRef, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { Bell, X } from 'lucide-react'
import { useAppState } from '../lib/app-state'
import { cn, formatDateTime, formatNotification } from '../lib/format'
import type { ApiNotification } from '../types/catalog'

const pollInterval = 15_000

const memory = {
  userId: '',
  items: [] as ApiNotification[],
  seen: null as Set<string> | null,
  readIds: new Set<string>(),
  toast: null as ApiNotification | null,
}

export function NotificationBell() {
  const { user, loadNotifications, markNotificationsRead } = useAppState()
  const userId = user?.id ?? ''
  if (memory.userId !== userId) {
    memory.userId = userId
    memory.items = []
    memory.seen = null
    memory.readIds = new Set()
    memory.toast = null
  }
  const [items, setItems] = useState<ApiNotification[]>(memory.items)
  const [open, setOpen] = useState(false)
  const [toast, setToast] = useState<ApiNotification | null>(memory.toast)
  const rootRef = useRef<HTMLDivElement>(null)
  const location = useLocation()
  const panelId = useId()
  const unread = items.some((item) => !item.read)

  function showToast(value: ApiNotification | null) {
    memory.toast = value
    setToast(value)
  }

  useEffect(() => {
    let active = true
    function refresh() {
      loadNotifications()
        .then((loaded) => {
          if (!active) return
          const next = loaded.map((item) => (memory.readIds.has(item.id) ? { ...item, read: true } : item))
          const known = memory.seen
          const fresh = known ? next.find((item) => !item.read && !known.has(item.id)) : undefined
          if (fresh) showToast(fresh)
          memory.seen = new Set(next.map((item) => item.id))
          memory.items = next
          setItems(next)
        })
        .catch(() => undefined)
    }
    refresh()
    const timer = window.setInterval(refresh, pollInterval)
    return () => {
      active = false
      window.clearInterval(timer)
    }
  }, [loadNotifications])

  useEffect(() => {
    if (!toast) return
    const timer = window.setTimeout(() => showToast(null), 5000)
    return () => window.clearTimeout(timer)
  }, [toast])

  useEffect(() => {
    setOpen(false)
  }, [location])

  useEffect(() => {
    if (!open) return
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') setOpen(false)
    }
    function onPointerDown(event: MouseEvent) {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false)
    }
    document.addEventListener('keydown', onKeyDown)
    document.addEventListener('mousedown', onPointerDown)
    return () => {
      document.removeEventListener('keydown', onKeyDown)
      document.removeEventListener('mousedown', onPointerDown)
    }
  }, [open])

  function toggle() {
    const next = !open
    setOpen(next)
    if (!next || !unread) return
    items.forEach((item) => memory.readIds.add(item.id))
    memory.items = items.map((item) => ({ ...item, read: true }))
    setItems(memory.items)
    showToast(null)
    void markNotificationsRead().catch(() => undefined)
  }

  return (
    <div ref={rootRef} className="relative">
      <button
        type="button"
        className={cn(
          'inline-flex size-11 items-center justify-center rounded-md hover:bg-ink-100',
          unread ? 'text-amber-500' : 'text-ink-900',
        )}
        aria-expanded={open}
        aria-controls={panelId}
        aria-label={unread ? 'Notifications, new updates' : 'Notifications'}
        onClick={toggle}
      >
        <Bell aria-hidden="true" className={cn('size-5', unread && 'fill-current')} />
      </button>
      <div
        id={panelId}
        hidden={!open}
        className="absolute right-0 top-full z-50 mt-2 w-80 max-w-[calc(100vw-2rem)] rounded-xl border border-ink-200 bg-white shadow-card"
      >
        <p className="border-b border-ink-200 px-4 py-3 text-sm font-semibold text-ink-900">Notifications</p>
        {items.length === 0 ? (
          <p className="px-4 py-6 text-sm text-ink-700">No notifications yet.</p>
        ) : (
          <ul className="max-h-96 overflow-y-auto">
            {items.map((item) => (
              <li key={item.id} className="border-b border-ink-100 last:border-b-0">
                <Link to={`/orders/${item.orderId}`} className="block px-4 py-3 hover:bg-ink-50">
                  <span className="block text-sm text-ink-900">{formatNotification(item.kind, item.productName)}</span>
                  <span className="mt-0.5 block text-xs text-ink-500">{formatDateTime(item.createdAt)}</span>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </div>
      <div role="status" aria-live="polite" className="fixed right-4 top-20 z-50 w-80 max-w-[calc(100vw-2rem)]">
        {toast && (
          <div className="flex items-start gap-3 rounded-xl border border-ink-200 bg-white p-4 shadow-lift">
            <Bell aria-hidden="true" className="mt-0.5 size-5 shrink-0 text-accent-700" />
            <div className="min-w-0 flex-1">
              <p className="text-sm font-semibold text-ink-900">{formatNotification(toast.kind, toast.productName)}</p>
              <p className="mt-0.5 text-xs text-ink-500">{formatDateTime(toast.createdAt)}</p>
            </div>
            <button
              type="button"
              className="inline-flex size-8 shrink-0 items-center justify-center rounded-md text-ink-700 hover:bg-ink-100"
              aria-label="Dismiss"
              onClick={() => showToast(null)}
            >
              <X aria-hidden="true" className="size-4" />
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
