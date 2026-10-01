import { useEffect } from 'react'
import { BrowserRouter, Route, Routes, useLocation } from 'react-router-dom'
import { LandingPage } from './pages/LandingPage'
import { StubPage } from './pages/StubPage'

function ScrollToHash() {
  const { hash, pathname } = useLocation()

  useEffect(() => {
    if (!hash) {
      window.scrollTo(0, 0)
      return
    }
    const target = document.getElementById(decodeURIComponent(hash.slice(1)))
    target?.scrollIntoView()
  }, [hash, pathname])

  return null
}

export function App() {
  return (
    <BrowserRouter>
      <ScrollToHash />
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/login" element={<StubPage title="Log in" />} />
        <Route path="/register" element={<StubPage title="Sign up" />} />
        <Route path="/catalog" element={<StubPage title="Catalog" />} />
        <Route path="/orders" element={<StubPage title="Open orders" />} />
        <Route path="/orders/new" element={<StubPage title="Create an order" />} />
        <Route path="*" element={<StubPage title="Page not found" />} />
      </Routes>
    </BrowserRouter>
  )
}
