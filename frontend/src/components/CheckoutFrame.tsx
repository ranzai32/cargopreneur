import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { Container } from './Container'
import { Logo } from './Logo'

export function CheckoutFrame({ backTo, backLabel, children }: { backTo: string; backLabel: string; children: ReactNode }) {
  return (
    <div className="min-h-screen bg-ink-50 text-ink-800">
      <header className="border-b border-ink-200 bg-white">
        <Container className="flex h-16 items-center justify-between gap-4">
          <Logo />
          <Link to={backTo} className="text-sm font-semibold text-accent-800">{backLabel}</Link>
        </Container>
      </header>
      <main id="main" className="py-8 sm:py-12">
        {children}
      </main>
    </div>
  )
}
