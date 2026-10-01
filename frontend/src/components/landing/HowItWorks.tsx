import { ChartColumn, CircleCheck, PackagePlus, Search } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { Container } from '../Container'
import { Section } from '../Section'

const steps: Array<{ title: string; body: string; icon: LucideIcon }> = [
  {
    title: 'Find or create an order',
    body: 'Search open orders, or start one for a product your business needs.',
    icon: Search,
  },
  {
    title: 'Join with your quantity',
    body: 'Add only what you will use. You are not buying the whole minimum yourself.',
    icon: PackagePlus,
  },
  {
    title: 'Track progress to the minimum',
    body: 'Watch the pooled quantity, the price tier, and the days left.',
    icon: ChartColumn,
  },
  {
    title: 'Confirmed, or cancelled',
    body: 'If the minimum is reached, the order is purchased at that wholesale tier and delivery is split between the businesses that joined. If not, it is cancelled and you have no obligation.',
    icon: CircleCheck,
  },
]

export function HowItWorks() {
  return (
    <Section id="how-it-works" labelledBy="how-heading" className="border-y border-ink-200 bg-white">
      <Container>
        <h2
          id="how-heading"
          className="max-w-2xl font-serif text-3xl leading-tight text-balance text-ink-950 sm:text-4xl"
        >
          How an order works
        </h2>
        <p className="mt-4 max-w-2xl text-ink-700">
          Four steps. You can see the minimum, the price, and the deadline before you join.
        </p>
        <ol className="mt-10 grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
          {steps.map((step, index) => (
            <li key={step.title}>
              <step.icon aria-hidden="true" className="size-5 text-accent-800" />
              <p className="mt-4 font-serif text-lg text-accent-800" aria-hidden="true">
                {String(index + 1).padStart(2, '0')}
              </p>
              <h3 className="mt-1 font-serif text-xl text-ink-950">{step.title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-ink-700">{step.body}</p>
            </li>
          ))}
        </ol>
      </Container>
    </Section>
  )
}
