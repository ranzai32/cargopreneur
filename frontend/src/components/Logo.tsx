export function Logo({ className = '' }: { className?: string }) {
  return (
    <span className={`inline-flex items-center gap-2.5 text-ink-950 ${className}`}>
      <svg
        viewBox="0 0 32 32"
        className="size-7 shrink-0 text-accent-800"
        aria-hidden="true"
      >
        <path
          d="M5 13.2 16 7.5l11 5.7v11.1L16 30.5 5 24.3V13.2Z"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.6"
        />
        <path
          d="M5 13.2 16 18.8 27 13.2M16 18.8v11.7"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.6"
        />
      </svg>
      <span className="font-serif text-xl leading-none tracking-tight">CargoPrener</span>
    </span>
  )
}
