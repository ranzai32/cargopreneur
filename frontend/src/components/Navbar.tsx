import { useEffect, useId, useRef, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { Menu, X } from 'lucide-react'
import { Button } from './Button'
import { Container } from './Container'
import { Logo } from './Logo'

const links = [
  { href: '/#how-it-works', label: 'How it works' },
  { href: '/catalog', label: 'Catalog' },
  { href: '/orders', label: 'Open orders' },
]

export function Navbar() {
  const [open, setOpen] = useState(false)
  const location = useLocation()
  const menuId = useId()
  const headerRef = useRef<HTMLElement>(null)
  const toggleRef = useRef<HTMLButtonElement>(null)
  const panelRef = useRef<HTMLDivElement>(null)
  const restoreFocus = useRef(false)

  useEffect(() => {
    setOpen(false)
  }, [location])

  useEffect(() => {
    if (!open && restoreFocus.current) {
      toggleRef.current?.focus()
      restoreFocus.current = false
    }
  }, [open])

  useEffect(() => {
    if (!open) return

    const first = panelRef.current?.querySelector<HTMLElement>('a, button')
    first?.focus()

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        event.preventDefault()
        restoreFocus.current = true
        setOpen(false)
        return
      }
      if (event.key !== 'Tab' || !panelRef.current) return
      const items = [
        ...panelRef.current.querySelectorAll<HTMLElement>('a[href], button:not([disabled])'),
      ]
      if (items.length === 0) return
      const firstItem = items[0]
      const lastItem = items[items.length - 1]
      if (event.shiftKey && document.activeElement === firstItem) {
        event.preventDefault()
        lastItem.focus()
      } else if (!event.shiftKey && document.activeElement === lastItem) {
        event.preventDefault()
        firstItem.focus()
      }
    }

    function onPointerDown(event: MouseEvent) {
      if (!headerRef.current?.contains(event.target as Node)) {
        setOpen(false)
      }
    }

    document.addEventListener('keydown', onKeyDown)
    document.addEventListener('mousedown', onPointerDown)
    return () => {
      document.removeEventListener('keydown', onKeyDown)
      document.removeEventListener('mousedown', onPointerDown)
    }
  }, [open])

  useEffect(() => {
    const media = window.matchMedia('(min-width: 1024px)')
    function onChange() {
      if (media.matches) setOpen(false)
    }
    media.addEventListener('change', onChange)
    return () => media.removeEventListener('change', onChange)
  }, [])

  function closeFromToggle() {
    if (open) restoreFocus.current = true
    setOpen((value) => !value)
  }

  return (
    <header ref={headerRef} className="sticky top-0 z-40 border-b border-ink-200 bg-ink-50">
      <Container className="flex h-16 items-center justify-between gap-4">
        <Link to="/" className="rounded-md">
          <Logo />
        </Link>
        <nav aria-label="Primary" className="hidden items-center gap-6 lg:flex">
          {links.map((link) => (
            <Link
              key={link.href}
              to={link.href}
              className="rounded-sm text-sm font-semibold text-ink-800 hover:text-accent-800"
            >
              {link.label}
            </Link>
          ))}
        </nav>
        <div className="hidden items-center gap-2 lg:flex">
          <Button to="/login" variant="ghost" size="sm">
            Log in
          </Button>
          <Button to="/register" size="sm">
            Sign up
          </Button>
        </div>
        <button
          ref={toggleRef}
          type="button"
          className="inline-flex size-11 items-center justify-center rounded-md text-ink-900 hover:bg-ink-100 lg:hidden"
          aria-expanded={open}
          aria-controls={menuId}
          aria-label="Menu"
          onClick={closeFromToggle}
        >
          {open ? <X aria-hidden="true" className="size-5" /> : <Menu aria-hidden="true" className="size-5" />}
        </button>
      </Container>
      <div
        id={menuId}
        ref={panelRef}
        hidden={!open}
        className="border-t border-ink-200 bg-ink-50 lg:hidden"
      >
        <Container className="flex flex-col py-2">
          <nav aria-label="Mobile">
            {links.map((link) => (
              <Link
                key={link.href}
                to={link.href}
                className="block rounded-md px-2 py-3 text-base font-semibold text-ink-900 hover:bg-ink-100"
              >
                {link.label}
              </Link>
            ))}
          </nav>
          <div className="mt-2 flex flex-col gap-2 border-t border-ink-200 py-4">
            <Button to="/login" variant="secondary" className="w-full">
              Log in
            </Button>
            <Button to="/register" className="w-full">
              Sign up
            </Button>
          </div>
        </Container>
      </div>
    </header>
  )
}
