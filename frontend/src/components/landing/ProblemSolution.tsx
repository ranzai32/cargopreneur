import { Container } from '../Container'
import { Section } from '../Section'

const problems = [
  'The supplier’s minimum is more than one shop or workshop can store or sell.',
  'Buying a smaller amount means paying a retail price.',
  'The purchase is often skipped, and the business goes without the product.',
]

const solutions = [
  'You add only your own quantity to an order other businesses can join.',
  'The unit price follows the pooled total, in published tiers.',
  'If the minimum is missed by the deadline, the order is cancelled.',
]

export function ProblemSolution() {
  return (
    <Section labelledBy="problem-heading">
      <Container>
        <h2
          id="problem-heading"
          className="max-w-2xl font-serif text-3xl leading-tight text-balance text-ink-950 sm:text-4xl"
        >
          Too much for one business. Enough when the order is shared.
        </h2>
        <div className="mt-10 grid gap-12 lg:grid-cols-2 lg:gap-16">
          <div>
            <h3 className="font-serif text-2xl text-ink-950">The problem</h3>
            <p className="mt-3 text-ink-700">
              Suppliers set a minimum order quantity. For a single small business,
              that number is often out of reach.
            </p>
            <ul className="mt-6 grid gap-4">
              {problems.map((item) => (
                <li key={item} className="border-l-2 border-ink-300 pl-4 text-ink-800">
                  {item}
                </li>
              ))}
            </ul>
          </div>
          <div>
            <h3 className="font-serif text-2xl text-ink-950">The shared order</h3>
            <p className="mt-3 text-ink-700">
              CargoPrener pools quantities from businesses that do not need to know
              each other. Each one commits only to what it will use.
            </p>
            <ul className="mt-6 grid gap-4">
              {solutions.map((item) => (
                <li key={item} className="border-l-2 border-accent-700 pl-4 text-ink-800">
                  {item}
                </li>
              ))}
            </ul>
          </div>
        </div>
      </Container>
    </Section>
  )
}
