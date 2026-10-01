import { Container } from '../Container'
import { Section } from '../Section'

const protections = [
  {
    title: 'Suppliers are reviewed first',
    body: 'An admin reviews a supplier before that supplier’s products can appear in the catalog.',
  },
  {
    title: 'Orders can be held',
    body: 'An admin can pause or remove an order if the product, the price, or the deadline is unclear or wrong.',
  },
  {
    title: 'Disputes have a reviewer',
    body: 'After a confirmed order, any business on it can open a dispute about delivery or the goods. An admin reviews the case and records the outcome.',
  },
]

export function Trust() {
  return (
    <Section id="trust" labelledBy="trust-heading">
      <Container>
        <h2
          id="trust-heading"
          className="max-w-2xl font-serif text-3xl leading-tight text-balance text-ink-950 sm:text-4xl"
        >
          How the platform protects an order
        </h2>
        <p className="mt-4 max-w-2xl text-ink-700">
          Moderation and disputes are handled by an admin. There is no automatic promise
          beyond the rules shown on the order.
        </p>
        <ol className="mt-10 grid gap-8 lg:grid-cols-3">
          {protections.map((item, index) => (
            <li key={item.title} className="border-t border-ink-300 pt-4">
              <p className="font-serif text-lg text-accent-800" aria-hidden="true">
                {index + 1}
              </p>
              <h3 className="mt-2 font-serif text-xl text-ink-950">{item.title}</h3>
              <p className="mt-2 text-ink-700">{item.body}</p>
            </li>
          ))}
        </ol>
      </Container>
    </Section>
  )
}
