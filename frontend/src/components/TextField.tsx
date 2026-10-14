import type { InputHTMLAttributes, ReactNode, SelectHTMLAttributes } from 'react'
import { cn } from '../lib/format'

const controlClass =
  'h-11 w-full rounded-md border border-ink-300 bg-white px-3 text-base text-ink-900'

export function TextField({
  id,
  label,
  error,
  ...props
}: {
  id: string
  label: string
  error?: string
} & InputHTMLAttributes<HTMLInputElement>) {
  return (
    <label className="block" htmlFor={id}>
      <span className="text-sm font-semibold text-ink-900">{label}</span>
      <input
        id={id}
        aria-invalid={Boolean(error)}
        aria-describedby={error ? `${id}-error` : undefined}
        className={cn(controlClass, 'mt-1.5', error && 'border-danger')}
        {...props}
      />
      {error && (
        <span id={`${id}-error`} className="mt-1 block text-sm text-danger">
          {error}
        </span>
      )}
    </label>
  )
}

export function SelectField({
  id,
  label,
  error,
  children,
  ...props
}: {
  id: string
  label: string
  error?: string
  children: ReactNode
} & SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <label className="block" htmlFor={id}>
      <span className="text-sm font-semibold text-ink-900">{label}</span>
      <select
        id={id}
        aria-invalid={Boolean(error)}
        aria-describedby={error ? `${id}-error` : undefined}
        className={cn(controlClass, 'mt-1.5', error && 'border-danger')}
        {...props}
      >
        {children}
      </select>
      {error && (
        <span id={`${id}-error`} className="mt-1 block text-sm text-danger">
          {error}
        </span>
      )}
    </label>
  )
}
