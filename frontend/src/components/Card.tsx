import type { ReactNode } from 'react'
import { cn } from '../lib/format'

type CardProps = {
  children: ReactNode
  className?: string
}

export function Card({ children, className }: CardProps) {
  return (
    <div
      className={cn(
        'rounded-xl border border-ink-200 bg-white shadow-card',
        className,
      )}
    >
      {children}
    </div>
  )
}
