export function PageHeader({ title, lede }: { title: string; lede?: string }) {
  return (
    <div className="max-w-2xl">
      <h1 className="font-serif text-4xl leading-tight text-ink-950">{title}</h1>
      {lede && <p className="mt-3 text-ink-700">{lede}</p>}
    </div>
  )
}
