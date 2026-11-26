import { useEffect } from 'react'
import { BrowserRouter, Route, Routes, useLocation } from 'react-router-dom'
import { AppStateProvider } from './lib/app-state'
import { AccountPage } from './pages/AccountPage'
import { AdminOrdersPage } from './pages/admin/OrdersPage'
import { AdminProductsPage } from './pages/admin/ProductsPage'
import { AdminSuppliersPage } from './pages/admin/SuppliersPage'
import { CatalogPage } from './pages/CatalogPage'
import { CreateOrderPage } from './pages/CreateOrderPage'
import { LandingPage } from './pages/LandingPage'
import { LoginPage } from './pages/LoginPage'
import { OrderPage } from './pages/OrderPage'
import { OrdersPage } from './pages/OrdersPage'
import { ProductPage } from './pages/ProductPage'
import { RegisterPage } from './pages/RegisterPage'
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
      <AppStateProvider>
        <ScrollToHash />
        <Routes>
          <Route path="/" element={<LandingPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/account" element={<AccountPage />} />
          <Route path="/catalog" element={<CatalogPage />} />
          <Route path="/catalog/:productId" element={<ProductPage />} />
          <Route path="/orders" element={<OrdersPage />} />
          <Route path="/orders/new" element={<CreateOrderPage />} />
          <Route path="/orders/:orderId" element={<OrderPage />} />
          <Route path="/admin/suppliers" element={<AdminSuppliersPage />} />
          <Route path="/admin/products" element={<AdminProductsPage />} />
          <Route path="/admin/orders" element={<AdminOrdersPage />} />
          <Route path="*" element={<StubPage title="Page not found" />} />
        </Routes>
      </AppStateProvider>
    </BrowserRouter>
  )
}
