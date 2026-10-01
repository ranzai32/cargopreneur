import { Link } from 'react-router-dom'
import { Container } from './Container'
import { Logo } from './Logo'

const groups = [
  {
    title: 'Platform',
    links: [
      { href: '/catalog', label: 'Catalog' },
      { href: '/orders', label: 'Open orders' },
      { href: '/#how-it-works', label: 'How it works' },
    ],
  },
  {
    title: 'Company',
    links: [
      { href: '/#trust', label: 'How orders stay safe' },
      { href: '/orders/new', label: 'Create an order' },
    ],
  },
  {
    title: 'Support',
    links: [
      { href: '/#example', label: 'Example order' },
      { href: '/register', label: 'Sign up' },
    ],
  },
]

export function Footer() {
  return (
    <footer className="border-t border-ink-200 bg-white">
      <Container className="py-12">
        <div className="grid gap-10 sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <Link to="/" className="rounded-md">
              <Logo />
            </Link>
            <p className="mt-4 max-w-xs text-sm leading-relaxed text-ink-700">
              CargoPrener helps small businesses pool orders and reach wholesale prices together.
            </p>
          </div>
          {groups.map((group) => (
            <div key={group.title}>
              <p className="text-sm font-semibold text-ink-950">{group.title}</p>
              <ul className="mt-3 grid gap-2">
                {group.links.map((link) => (
                  <li key={link.href}>
                    <Link
                      to={link.href}
                      className="rounded-sm text-sm text-ink-700 hover:text-accent-800"
                    >
                      {link.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
        <p className="mt-10 border-t border-ink-200 pt-6 text-sm text-ink-600">
          © {new Date().getFullYear()} CargoPrener
        </p>
      </Container>
    </footer>
  )
}
