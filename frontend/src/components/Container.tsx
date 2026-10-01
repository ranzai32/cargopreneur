import type { ReactNode } from 'react'
import { cn } from '../lib/format'

type ContainerProps = {
  children: ReactNode
  className?: string
}

export function Container({ children, className }: ContainerProps) {
  return (
    <div className={cn('mx-auto w-full max-w-6xl px-5 sm:px-6', className)}>
      {children}
    </div>
  )
}
