import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { Button } from '../components/Button'
import { Card } from '../components/Card'
import { Container } from '../components/Container'
import { PageLayout } from '../components/PageLayout'
import { TextField } from '../components/TextField'
import { fieldErrorsOf, isApiError, useAppState } from '../lib/app-state'

export function LoginPage() {
  const { login } = useAppState()
  const navigate = useNavigate()
  const location = useLocation()
  const from = (location.state as { from?: string } | null)?.from ?? '/catalog'
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [message, setMessage] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const nextErrors: Record<string, string> = {}
    if (!email.trim()) nextErrors.email = 'Email is required'
    if (!password) nextErrors.password = 'Password is required'
    setFieldErrors(nextErrors)
    setMessage('')
    if (Object.keys(nextErrors).length > 0) return
    setPending(true)
    try {
      await login(email, password)
      navigate(from)
    } catch (error) {
      setFieldErrors(fieldErrorsOf(error))
      setMessage(isApiError(error) ? error.message : 'Could not log in')
    } finally {
      setPending(false)
    }
  }

  return (
    <PageLayout>
      <Container className="py-section">
        <Card className="mx-auto max-w-md p-6 sm:p-8">
          <h1 className="font-serif text-3xl text-ink-950">Log in</h1>
          <p className="mt-2 text-sm text-ink-700">Use the email for your business account.</p>
          <form className="mt-6 grid gap-4" onSubmit={handleSubmit} noValidate>
            {message && (
              <p role="alert" className="text-sm text-danger">
                {message}
              </p>
            )}
            <TextField id="email" label="Email" type="email" autoComplete="email" value={email} error={fieldErrors.email} onChange={(event) => setEmail(event.target.value)} />
            <TextField id="password" label="Password" type="password" autoComplete="current-password" value={password} error={fieldErrors.password} onChange={(event) => setPassword(event.target.value)} />
            <Button type="submit" disabled={pending}>
              {pending ? 'Logging in…' : 'Log in'}
            </Button>
          </form>
          <p className="mt-4 text-sm text-ink-700">
            New here? <Link to="/register" className="font-semibold text-accent-800">Create an account</Link>
          </p>
        </Card>
      </Container>
    </PageLayout>
  )
}
