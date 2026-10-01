import type { ReactNode } from 'react'
import { cn } from '../lib/format'

type Tone = 'neutral' | 'accent' | 'success' | 'warning' | 'danger'

const toneClass: Record<Tone, string> = {
  neutral: 'bg-ink-100 text-ink-800',
  accent: 'bg-accent-100 text-accent-800',
  success: 'bg-success-soft text-success',
  warning: 'bg-warning-soft text-warning',
  danger: 'bg-danger-soft text-danger',
}

type BadgeProps = {
  children: ReactNode
  tone?: Tone
  className?: string
}

export function Badge({ children, tone = 'neutral', className }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-md px-2 py-0.5 text-xs font-semibold tracking-wide',
        toneClass[tone],
        className,
      )}
    >
      {children}
    </span>
  )
}
