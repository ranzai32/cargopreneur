import type { NotificationKind } from '../types/catalog'

const money = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
})

const number = new Intl.NumberFormat('en-US')

export function formatMoney(value: number): string {
  return money.format(value)
}

export function formatNumber(value: number): string {
  return number.format(value)
}

export function formatQuantity(count: number, unit: string): string {
  return `${formatNumber(count)} ${count === 1 ? unit : pluralUnit(unit)}`
}

function pluralUnit(unit: string) {
  if (unit.endsWith('s')) return unit
  if (/[sxz]$/.test(unit) || unit.endsWith('ch') || unit.endsWith('sh')) return `${unit}es`
  return `${unit}s`
}

export function daysUntil(isoDate: string, now = new Date()): number {
  const [year, month, day] = isoDate.split('-').map(Number)
  if (!year || !month || !day) return 0
  const end = new Date(year, month - 1, day, 23, 59, 59, 999)
  const diff = end.getTime() - now.getTime()
  if (diff <= 0) return 0
  return Math.ceil(diff / 86_400_000)
}

export function formatDeliveryStatus(status: 'pending' | 'shipped' | 'delivered' | null | undefined): string {
  if (status === 'shipped') return 'Shipped'
  if (status === 'delivered') return 'Delivered'
  return 'Pending'
}

export function formatReceiptStatus(status: 'pending' | 'received' | null | undefined): string {
  if (status === 'received') return 'Received'
  return 'Not confirmed'
}

export function formatDaysLeft(days: number): string {
  if (days <= 0) return 'Deadline passed'
  if (days === 1) return '1 day left'
  return `${days} days left`
}

export function formatNotification(kind: NotificationKind, productName: string): string {
  switch (kind) {
    case 'order_created':
      return `Your order for ${productName} was created. Other businesses can join it now.`
    case 'order_confirmed':
      return `Your order for ${productName} reached its minimum. You can pay now.`
    case 'order_cancelled':
      return `Your order for ${productName} was cancelled.`
    case 'payment_received':
      return `Payment received for ${productName}.`
    case 'shipment_sent':
      return `Your ${productName} shipment is on its way.`
    case 'shipment_delivered':
      return `Your ${productName} shipment was delivered.`
    case 'dispute_opened':
      return `A dispute was opened on the ${productName} order.`
    case 'dispute_resolved':
      return `Your dispute on the ${productName} order was resolved.`
  }
}

export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' })
}

export function cn(...parts: Array<string | false | null | undefined>): string {
  return parts.filter(Boolean).join(' ')
}
