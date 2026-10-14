import type { ReactNode } from 'react'
import { Footer } from './Footer'
import { ExampleNotice } from './ExampleNotice'
import { Navbar } from './Navbar'

export function PageLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col bg-ink-50 text-ink-800">
      <a href="#main" className="skip-link">
        Skip to content
      </a>
      <Navbar />
      <ExampleNotice />
      <main id="main" className="flex-1">
        {children}
      </main>
      <Footer />
    </div>
  )
}
