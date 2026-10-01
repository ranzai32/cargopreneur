import { Button } from '../Button'
import { Card } from '../Card'
import { Container } from '../Container'
import { Section } from '../Section'

export function FinalCta() {
  return (
    <Section labelledBy="cta-heading">
      <Container>
        <Card className="px-6 py-10 sm:px-10 sm:py-12">
          <div className="flex flex-col gap-8 md:flex-row md:items-end md:justify-between">
            <div className="max-w-xl">
              <h2
                id="cta-heading"
                className="font-serif text-3xl leading-tight text-balance text-ink-950 sm:text-4xl"
              >
                Join an order, or start one
              </h2>
              <p className="mt-4 text-ink-700">
                Look through open orders, or create an order for a product your business
                needs. You are not obliged to buy unless the minimum is reached.
              </p>
            </div>
            <div className="flex flex-col gap-3 sm:flex-row">
              <Button to="/orders" size="lg">
                Join an order
              </Button>
              <Button to="/orders/new" variant="secondary" size="lg">
                Create an order
              </Button>
            </div>
          </div>
        </Card>
      </Container>
    </Section>
  )
}
