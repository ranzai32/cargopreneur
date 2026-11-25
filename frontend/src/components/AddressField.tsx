import { useEffect, useId, useState } from 'react'
import { api } from '../lib/api'

export function AddressField({
  id,
  label,
  value,
  error,
  onChange,
}: {
  id: string
  label: string
  value: string
  error?: string
  onChange: (value: string) => void
}) {
  const listId = useId()
  const [items, setItems] = useState<string[]>([])

  useEffect(() => {
    const query = value.trim()
    if (query.length < 3) {
      setItems([])
      return
    }
    let cancel = false
    const timer = window.setTimeout(() => {
      api.addresses(query)
        .then((result) => {
          if (!cancel) setItems(result.items)
        })
        .catch(() => {
          if (!cancel) setItems([])
        })
    }, 300)
    return () => {
      cancel = true
      window.clearTimeout(timer)
    }
  }, [value])

  return (
    <div className="relative">
      <label className="block" htmlFor={id}>
        <span className="text-sm font-semibold text-ink-900">{label}</span>
        <input
          id={id}
          value={value}
          role="combobox"
          aria-expanded={items.length > 0}
          aria-controls={listId}
          aria-invalid={Boolean(error)}
          aria-describedby={error ? `${id}-error` : undefined}
          autoComplete="off"
          className={`mt-1.5 h-11 w-full rounded-md border bg-white px-3 text-base text-ink-900 ${error ? 'border-danger' : 'border-ink-300'}`}
          onChange={(event) => onChange(event.target.value)}
        />
      </label>
      {error && (
        <span id={`${id}-error`} className="mt-1 block text-sm text-danger">
          {error}
        </span>
      )}
      {items.length > 0 && (
        <ul id={listId} role="listbox" className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md border border-ink-300 bg-white py-1 shadow-sm">
          {items.map((item) => (
            <li key={item}>
              <button
                type="button"
                role="option"
                className="w-full px-3 py-2 text-left text-sm text-ink-900 hover:bg-ink-100"
                onClick={() => {
                  onChange(item)
                  setItems([])
                }}
              >
                {item}
              </button>
            </li>
          ))}
          <li className="px-3 py-1 text-xs text-ink-600">Suggestions by 2GIS</li>
        </ul>
      )}
    </div>
  )
}
