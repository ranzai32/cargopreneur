import { useEffect } from 'react'
import { BrowserRouter, Route, Routes, useLocation } from 'react-router-dom'
import { AppStateProvider } from './lib/app-state'
import { AccountPage } from './pages/AccountPage'
import { AdminDisputesPage } from './pages/admin/DisputesPage'
import { AdminOrdersPage } from './pages/admin/OrdersPage'
import { AdminProductsPage } from './pages/admin/ProductsPage'
import { AdminSuppliersPage } from './pages/admin/SuppliersPage'
import { CatalogPage } from './pages/CatalogPage'
import { CheckoutPage } from './pages/CheckoutPage'
import { CreateOrderPage } from './pages/CreateOrderPage'
import { LandingPage } from './pages/LandingPage'
import { LoginPage } from './pages/LoginPage'
import { OrderPage } from './pages/OrderPage'
import { OrdersPage } from './pages/OrdersPage'
import { PaymentProcessingPage } from './pages/PaymentProcessingPage'
import { ProductPage } from './pages/ProductPage'
import { ProductReviewsPage, SupplierReviewsPage } from './pages/ReviewsPage'
import { RegisterPage } from './pages/RegisterPage'
import { SupplierPage } from './pages/SupplierPage'
import { SupplierOrdersPage, SupplierProductsPage } from './pages/supplier/ShopPage'
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
          <Route path="/catalog/:productId/reviews" element={<ProductReviewsPage />} />
          <Route path="/catalog/:productId" element={<ProductPage />} />
          <Route path="/suppliers/:supplierId/reviews" element={<SupplierReviewsPage />} />
          <Route path="/suppliers/:supplierId" element={<SupplierPage />} />
          <Route path="/supplier" element={<SupplierProductsPage />} />
          <Route path="/supplier/orders" element={<SupplierOrdersPage />} />
          <Route path="/orders" element={<OrdersPage />} />
          <Route path="/orders/new" element={<CreateOrderPage />} />
          <Route path="/orders/:orderId/checkout" element={<CheckoutPage />} />
          <Route path="/orders/:orderId/processing" element={<PaymentProcessingPage />} />
          <Route path="/orders/:orderId" element={<OrderPage />} />
          <Route path="/admin/suppliers" element={<AdminSuppliersPage />} />
          <Route path="/admin/products" element={<AdminProductsPage />} />
          <Route path="/admin/orders" element={<AdminOrdersPage />} />
          <Route path="/admin/disputes" element={<AdminDisputesPage />} />
          <Route path="*" element={<StubPage title="Page not found" />} />
        </Routes>
      </AppStateProvider>
    </BrowserRouter>
  )
}
