import { useEffect, useState } from 'react'
import type { PriceTier } from '../types'
import { resolveTier } from '../lib/pricing'
import { cn, formatMoney, formatNumber } from '../lib/format'

type ProgressBarProps = {
  current: number
  target: number
  tiers?: PriceTier[]
  unit?: string
  showScaleNote?: boolean
}

export function ProgressBar({
  current,
  target,
  tiers = [],
  unit = '',
  showScaleNote = false,
}: ProgressBarProps) {
  const tierPeak = tiers.reduce((max, tier) => Math.max(max, tier.minQty), 0)
  const scale = Math.max(target, current, tierPeak, 1)
  const goal = Math.min(100, (Math.max(current, 0) / scale) * 100)
  const reducedAtStart = prefersReducedMotion()
  const [width, setWidth] = useState(reducedAtStart ? goal : 0)
  const active = resolveTier(current, tiers)
  const unitLabel = unit ? ` ${unit}` : ''

  useEffect(() => {
    if (prefersReducedMotion()) {
      setWidth(goal)
      return
    }
    const frame = requestAnimationFrame(() => setWidth(goal))
    return () => cancelAnimationFrame(frame)
  }, [goal])

  const minimumOnTier = tiers.some((tier) => tier.minQty === target)
  const sorted = [...tiers].sort((a, b) => a.minQty - b.minQty)

  return (
    <div>
      <div
        className="relative h-2.5 rounded-full bg-ink-100"
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={target}
        aria-valuenow={Math.min(Math.max(current, 0), target)}
        aria-valuetext={
          current >= target
            ? `${formatNumber(current)}${unitLabel} collected. Minimum of ${formatNumber(target)} is reached.`
            : `${formatNumber(current)} of ${formatNumber(target)}${unitLabel} collected.`
        }
        aria-label="Collected quantity toward the minimum"
      >
        <div
          className="progress-fill h-full rounded-full bg-accent-700"
          style={{ width: `${width}%` }}
        />
        {sorted.map((tier) => (
          <span
            key={tier.minQty}
            aria-hidden="true"
            className="absolute top-1/2 z-10 h-3.5 w-0.5 -translate-x-1/2 -translate-y-1/2 bg-ink-950 ring-2 ring-white"
            style={{ left: `${(tier.minQty / scale) * 100}%` }}
          />
        ))}
        {!minimumOnTier && target > 0 && (
          <span
            aria-hidden="true"
            className="absolute top-1/2 z-10 h-3.5 w-0.5 -translate-x-1/2 -translate-y-1/2 bg-ink-950 ring-2 ring-white"
            style={{ left: `${(target / scale) * 100}%` }}
          />
        )}
      </div>
      <ul className="mt-3 border-t border-ink-200">
        {!minimumOnTier && (
          <li className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1 border-b border-ink-200 py-2 text-sm">
            <span className="font-semibold text-ink-950">
              Minimum {formatNumber(target)}
            </span>
            <span className="text-ink-700">Required to confirm</span>
          </li>
        )}
        {sorted.map((tier) => {
          const isActive = active?.minQty === tier.minQty && active.price === tier.price
          const isMinimum = tier.minQty === target
          return (
            <li
              key={`${tier.minQty}-${tier.price}`}
              aria-current={isActive ? 'true' : undefined}
              className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1 border-b border-ink-200 py-2 text-sm"
            >
              <span className={cn('font-semibold tabular-nums', isActive ? 'text-accent-800' : 'text-ink-950')}>
                {formatNumber(tier.minQty)}+{isMinimum ? ' · Minimum' : ''}
                {isActive ? ' · Current' : ''}
              </span>
              <span className="shrink-0 tabular-nums text-ink-800">
                {formatMoney(tier.price)}
                {unit ? ` / ${unit}` : ''}
              </span>
            </li>
          )
        })}
      </ul>
      {showScaleNote && (
        <p className="mt-2 text-sm text-ink-700">
          Tiers above the minimum are marked further along the bar.
        </p>
      )}
    </div>
  )
}

function prefersReducedMotion(): boolean {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}
