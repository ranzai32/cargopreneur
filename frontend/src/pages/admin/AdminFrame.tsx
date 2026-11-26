import { useEffect } from 'react'
import type { ReactNode } from 'react'
import { NavLink } from 'react-router-dom'
import { Container } from '../../components/Container'
import { PageHeader } from '../../components/PageHeader'
import { PageLayout } from '../../components/PageLayout'
import { useAppState } from '../../lib/app-state'

const links = [
  { to: '/admin/suppliers', label: 'Suppliers' },
  { to: '/admin/products', label: 'Products' },
  { to: '/admin/orders', label: 'Orders' },
  { to: '/admin/disputes', label: 'Disputes' },
]

export function AdminFrame({
  title,
  lede,
  children,
}: {
  title: string
  lede: string
  children: ReactNode
}) {
  const { user, ready } = useAppState()
  const admin = user?.role === 'admin'

  useEffect(() => {
    document.title = `${title} · CargoPrener`
  }, [title])

  return (
    <PageLayout>
      <Container className="py-section">
        <PageHeader title={title} lede={lede} />
        {!ready && <p className="mt-8 text-ink-700">Loading…</p>}
        {ready && !admin && (
          <p role="alert" className="mt-8 text-danger">You do not have access to this</p>
        )}
        {ready && admin && (
          <>
            <nav aria-label="Admin" className="mt-6 flex flex-wrap gap-2">
              {links.map((link) => (
                <NavLink
                  key={link.to}
                  to={link.to}
                  className={({ isActive }) => (
                    isActive
                      ? 'rounded-md bg-accent-700 px-3 py-2 text-sm font-semibold text-white'
                      : 'rounded-md border border-ink-300 bg-white px-3 py-2 text-sm font-semibold text-ink-800'
                  )}
                >
                  {link.label}
                </NavLink>
              ))}
            </nav>
            <div className="mt-8">{children}</div>
          </>
        )}
      </Container>
    </PageLayout>
  )
}
