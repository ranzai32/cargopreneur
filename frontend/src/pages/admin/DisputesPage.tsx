import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../components/Button'
import { Card } from '../../components/Card'
import { isApiError, useAppState } from '../../lib/app-state'
import type { AdminDispute } from '../../types/catalog'
import { AdminFrame } from './AdminFrame'

export function AdminDisputesPage() {
  return (
    <AdminFrame
      title="Disputes"
      lede="Review a problem with delivery or the goods, then record the outcome."
    >
      <DisputeManager />
    </AdminFrame>
  )
}

function DisputeManager() {
  const { loadAdminDisputes, resolveDispute } = useAppState()
  const [disputes, setDisputes] = useState<AdminDispute[]>([])
  const [loaded, setLoaded] = useState(false)
  const [outcomes, setOutcomes] = useState<Record<string, string>>({})
  const [message, setMessage] = useState('')
  const [updating, setUpdating] = useState('')
  const [blocked, setBlocked] = useState('')

  useEffect(() => {
    let cancel = false
    loadAdminDisputes()
      .then((items) => {
        if (!cancel) {
          setDisputes(items)
          setLoaded(true)
        }
      })
      .catch((error) => {
        if (cancel) return
        if (isApiError(error) && (error.status === 401 || error.status === 403)) {
          setBlocked(error.status === 403 ? 'You do not have access to this' : error.message)
          return
        }
        setMessage(isApiError(error) ? error.message : 'Disputes could not be loaded')
        setLoaded(true)
      })
    return () => {
      cancel = true
    }
  }, [loadAdminDisputes])

  async function handleResolve(id: string) {
    setUpdating(id)
    setMessage('')
    try {
      const resolved = await resolveDispute(id, outcomes[id] ?? '')
      setDisputes((current) => current.map((item) => (item.id === id ? resolved : item)))
      setOutcomes((current) => ({ ...current, [id]: '' }))
    } catch (error) {
      if (isApiError(error) && (error.status === 401 || error.status === 403)) {
        setBlocked(error.status === 403 ? 'You do not have access to this' : error.message)
        return
      }
      const fields = isApiError(error) ? error.fieldErrors : undefined
      setMessage(fields?.outcome || (isApiError(error) ? error.message : 'Could not resolve the dispute'))
    } finally {
      setUpdating('')
    }
  }

  if (blocked) {
    return <p role="alert" className="text-danger">{blocked}</p>
  }

  return (
    <div>
      {message && <p role="alert" className="mb-4 text-danger">{message}</p>}
      {!loaded && <p className="text-ink-700">Loading disputes…</p>}
      {loaded && disputes.length === 0 && <p className="text-ink-700">No disputes yet.</p>}
      <ul className="grid gap-4">
        {disputes.map((dispute) => (
          <li key={dispute.id}>
            <Card className="p-5 sm:p-6">
              <p className="text-sm font-semibold text-ink-900">{dispute.status === 'open' ? 'Open' : 'Resolved'}</p>
              <h2 className="mt-2 font-serif text-2xl text-ink-950">
                <Link to={`/orders/${dispute.orderId}`} className="rounded-sm hover:text-accent-800">{dispute.productName}</Link>
              </h2>
              <p className="mt-1 text-sm text-ink-700">{dispute.businessName} · {dispute.orderStatus}</p>
              <p className="mt-3 text-ink-800">{dispute.reason}</p>
              {dispute.status === 'open' ? (
                <div className="mt-4">
                  <label className="block text-sm font-semibold text-ink-900" htmlFor={`outcome-${dispute.id}`}>Outcome</label>
                  <textarea
                    id={`outcome-${dispute.id}`}
                    value={outcomes[dispute.id] ?? ''}
                    rows={3}
                    className="mt-1.5 w-full rounded-md border border-ink-300 bg-white px-3 py-2 text-base text-ink-900"
                    onChange={(event) => setOutcomes((current) => ({ ...current, [dispute.id]: event.target.value }))}
                  />
                  <Button className="mt-3" size="sm" disabled={updating === dispute.id} onClick={() => void handleResolve(dispute.id)}>
                    {updating === dispute.id ? 'Saving…' : 'Resolve'}
                  </Button>
                </div>
              ) : (
                <p className="mt-3 text-sm text-ink-700">Outcome: {dispute.outcome}</p>
              )}
            </Card>
          </li>
        ))}
      </ul>
    </div>
  )
}
