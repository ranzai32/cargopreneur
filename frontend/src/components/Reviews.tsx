import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Button } from './Button'
import { isApiError } from '../lib/app-state'
import type { ApiReview } from '../types/catalog'

export function Stars({ value }: { value: number }) {
  const score = Math.max(0, Math.min(5, value))
  return (
    <span className="inline-flex items-center gap-0.5" aria-label={`${score.toFixed(1)} out of 5 stars`}>
      {[0, 1, 2, 3, 4].map((index) => {
        const fill = Math.max(0, Math.min(1, score - index))
        return (
          <span key={index} className="relative inline-block h-4 w-4 text-ink-300">
            <StarGlyph />
            <span className="absolute inset-y-0 left-0 overflow-hidden text-amber-500" style={{ width: `${fill * 100}%` }}>
              <StarGlyph />
            </span>
          </span>
        )
      })}
    </span>
  )
}

function StarGlyph() {
  return (
    <svg viewBox="0 0 20 20" className="h-4 w-4" aria-hidden="true">
      <path fill="currentColor" d="M10 1.6 12.5 7l5.9.5-4.5 3.8 1.4 5.8L10 14.6 4.7 17.1l1.4-5.8L1.6 7.5 7.5 7z" />
    </svg>
  )
}

export function RatingLink({ to, rating = 0, count = 0 }: { to: string; rating?: number; count?: number }) {
  return (
    <Link to={to} className="inline-flex items-center gap-2 text-sm text-ink-800">
      <Stars value={count ? rating : 0} />
      <span>{count ? `${rating.toFixed(1)} · ${count}` : 'No reviews'}</span>
    </Link>
  )
}

export function ReviewList({ reviews }: { reviews: ApiReview[] }) {
  if (reviews.length === 0) {
    return <p className="mt-3 text-sm text-ink-700">No reviews yet.</p>
  }
  return (
    <ul className="mt-4 grid gap-3">
      {reviews.map((review) => (
        <li key={review.id} className="rounded-md border border-ink-200 p-3">
          <p className="flex flex-wrap items-center gap-2 text-sm font-semibold text-ink-900">
            <span>{review.author}</span>
            <Stars value={review.rating} />
          </p>
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
