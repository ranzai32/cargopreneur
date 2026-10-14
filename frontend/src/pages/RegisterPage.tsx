import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button } from '../components/Button'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageLayout } from '../components/PageLayout'
import { TextField } from '../components/TextField'
import { fieldErrorsOf, isApiError, useAppState } from '../lib/app-state'

const empty = { businessName: '', contactName: '', email: '', password: '', region: '' }

export function RegisterPage() {
  const { register } = useAppState()
  const navigate = useNavigate()
  const [form, setForm] = useState(empty)
  const [message, setMessage] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)

  function update(key: keyof typeof empty, value: string) {
    setForm((current) => ({ ...current, [key]: value }))
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const nextErrors: Record<string, string> = {}
    if (!form.businessName.trim()) nextErrors.businessName = 'Business name is required'
    if (!form.contactName.trim()) nextErrors.contactName = 'Contact name is required'
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim())) nextErrors.email = 'Enter a valid email address'
    if (form.password.length < 8) nextErrors.password = 'Password must be at least 8 characters'
    if (!form.region.trim()) nextErrors.region = 'Region is required'
    setFieldErrors(nextErrors)
    setMessage('')
    if (Object.keys(nextErrors).length > 0) return
    setPending(true)
    try {
      await register({
        businessName: form.businessName.trim(),
        contactName: form.contactName.trim(),
        email: form.email.trim(),
        password: form.password,
        region: form.region.trim(),
      })
      navigate('/catalog')
    } catch (error) {
      setFieldErrors(fieldErrorsOf(error))
      setMessage(isApiError(error) ? error.message : 'Could not create the account')
    } finally {
      setPending(false)
    }
  }

  return (
    <PageLayout>
      <Container className="py-section">
        <Card className="mx-auto max-w-md p-6 sm:p-8">
          <h1 className="font-serif text-3xl text-ink-950">Create an account</h1>
          <p className="mt-2 text-sm text-ink-700">One account for your shop or workshop.</p>
          <form className="mt-6 grid gap-4" onSubmit={handleSubmit} noValidate>
            {message && (
              <p role="alert" className="text-sm text-danger">
                {message}
              </p>
            )}
            <TextField id="businessName" label="Business name" autoComplete="organization" value={form.businessName} error={fieldErrors.businessName} onChange={(event) => update('businessName', event.target.value)} />
            <TextField id="contactName" label="Your name" autoComplete="name" value={form.contactName} error={fieldErrors.contactName} onChange={(event) => update('contactName', event.target.value)} />
            <TextField id="email" label="Email" type="email" autoComplete="email" value={form.email} error={fieldErrors.email} onChange={(event) => update('email', event.target.value)} />
            <TextField id="region" label="City and region" autoComplete="address-level1" value={form.region} error={fieldErrors.region} onChange={(event) => update('region', event.target.value)} />
            <TextField id="password" label="Password" type="password" autoComplete="new-password" value={form.password} error={fieldErrors.password} onChange={(event) => update('password', event.target.value)} />
            <Button type="submit" disabled={pending}>
              {pending ? 'Creating account…' : 'Create account'}
            </Button>
          </form>
          <p className="mt-4 text-sm text-ink-700">
            Already registered? <Link to="/login" className="font-semibold text-accent-800">Log in</Link>
          </p>
        </Card>
      </Container>
    </PageLayout>
  )
}
