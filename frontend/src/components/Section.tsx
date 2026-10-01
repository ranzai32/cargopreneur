import type { ReactNode } from 'react'
import { cn } from '../lib/format'

type SectionProps = {
  id?: string
  children: ReactNode
  className?: string
  labelledBy?: string
}

export function Section({ id, children, className, labelledBy }: SectionProps) {
  return (
    <section
      id={id}
      aria-labelledby={labelledBy}
      className={cn('scroll-mt-20 py-section md:py-section-lg', className)}
    >
      {children}
    </section>
  )
}
