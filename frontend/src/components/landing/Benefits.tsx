import { BadgeCheck, Eye, ShieldCheck, Wallet } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { Card } from '../Card'
import { Container } from '../Container'
import { Section } from '../Section'

const benefits: Array<{ title: string; body: string; icon: LucideIcon }> = [
  {
    title: 'Save money',
    body: 'Each product has wholesale tiers. The unit price drops when the pooled quantity reaches the next tier. You can see that price before you join.',
    icon: Wallet,
  },
  {
    title: 'See progress in real time',
    body: 'An open order shows how much is collected, how much is still needed, and how many days are left. Those figures update as other businesses join.',
    icon: Eye,
  },
  {
    title: 'Safe commitment',
    body: 'The minimum, the price tiers, and the deadline are on the order before you join. If the minimum is not reached, the order is cancelled and you have no obligation to buy.',
    icon: ShieldCheck,
  },
  {
    title: 'Moderated suppliers',
    body: 'Suppliers are reviewed before their products can be listed. You are not joining an anonymous posting with no check.',
    icon: BadgeCheck,
  },
]

export function Benefits() {
  return (
    <Section labelledBy="benefits-heading" className="border-y border-ink-200 bg-white">
      <Container>
        <h2
          id="benefits-heading"
          className="max-w-2xl font-serif text-3xl leading-tight text-balance text-ink-950 sm:text-4xl"
        >
          What you can count on
        </h2>
        <div className="mt-10 grid gap-4 sm:grid-cols-2">
          {benefits.map((benefit) => (
            <Card key={benefit.title} className="p-6">
              <benefit.icon aria-hidden="true" className="size-5 text-accent-800" />
              <h3 className="mt-4 font-serif text-2xl text-ink-950">{benefit.title}</h3>
              <p className="mt-2 text-ink-700">{benefit.body}</p>
            </Card>
          ))}
        </div>
      </Container>
    </Section>
  )
}
