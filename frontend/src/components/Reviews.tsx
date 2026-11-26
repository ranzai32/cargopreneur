import { useState } from 'react'
import type { FormEvent } from 'react'
import { Button } from './Button'
import { isApiError } from '../lib/app-state'
import type { ApiReview } from '../types/catalog'

export function formatRating(rating?: number, count?: number) {
  if (!count) return 'No reviews yet'
  return `${(rating ?? 0).toFixed(1)} · ${count} ${count === 1 ? 'review' : 'reviews'}`
}

export function ReviewList({ reviews }: { reviews: ApiReview[] }) {
  if (reviews.length === 0) {
    return <p className="mt-3 text-sm text-ink-700">No reviews yet.</p>
  }
  return (
    <ul className="mt-4 grid gap-3">
      {reviews.map((review) => (
        <li key={review.id} className="rounded-md border border-ink-200 p-3">
          <p className="text-sm font-semibold text-ink-900">{review.author} · {review.rating}/5</p>
          <p className="mt-1 text-sm text-ink-800">{review.body}</p>
        </li>
      ))}
    </ul>
  )
}

export function ReviewForm({ onSubmit }: { onSubmit: (rating: number, body: string) => Promise<void> }) {
  const [rating, setRating] = useState('5')
  const [body, setBody] = useState('')
  const [message, setMessage] = useState('')
  const [pending, setPending] = useState(false)

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setPending(true)
    setMessage('')
    try {
      await onSubmit(Number(rating), body)
      setBody('')
      setMessage('Review saved.')
    } catch (error) {
      setMessage(isApiError(error) ? error.message : 'Could not save the review')
    } finally {
      setPending(false)
    }
  }

  return (
    <form className="mt-4 grid gap-3" onSubmit={handleSubmit}>
      {message && <p className="text-sm text-ink-800">{message}</p>}
      <label className="block text-sm font-semibold text-ink-900">
        Rating
        <select value={rating} onChange={(event) => setRating(event.target.value)} className="mt-1.5 h-11 w-full rounded-md border border-ink-300 bg-white px-3">
          {[5, 4, 3, 2, 1].map((value) => (
            <option key={value} value={value}>{value}</option>
          ))}
        </select>
      </label>
      <label className="block text-sm font-semibold text-ink-900">
        Review
        <textarea value={body} rows={3} onChange={(event) => setBody(event.target.value)} className="mt-1.5 w-full rounded-md border border-ink-300 bg-white px-3 py-2" />
      </label>
      <Button type="submit" disabled={pending}>{pending ? 'Saving…' : 'Save review'}</Button>
    </form>
  )
}
