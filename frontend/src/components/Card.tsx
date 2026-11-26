import type { ReactNode } from 'react'
import { cn } from '../lib/format'

type CardProps = {
  children: ReactNode
  className?: string
  id?: string
}

export function Card({ children, className, id }: CardProps) {
  return (
    <div
      id={id}
      className={cn(
        'rounded-xl border border-ink-200 bg-white shadow-card',
        className,
      )}
    >
      {children}
    </div>
  )
}
