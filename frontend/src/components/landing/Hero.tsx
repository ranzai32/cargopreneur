import { Button } from '../Button'
import { Card } from '../Card'
import { Container } from '../Container'
import { OrderPreview } from '../OrderPreview'
import { exampleOrder, exampleProduct, exampleSupplier } from '../../mocks/exampleOrder'

export function Hero() {
  return (
    <section className="py-section md:py-section-lg">
      <Container>
        <div className="grid items-center gap-12 lg:grid-cols-[minmax(0,1.05fr)_minmax(0,0.95fr)] lg:gap-16">
          <div>
            <h1 className="max-w-xl font-serif text-4xl leading-[1.12] text-balance text-ink-950 sm:text-5xl lg:text-6xl">
              Buy together. Pay the wholesale price.
            </h1>
            <p className="mt-5 max-w-xl text-lg leading-relaxed text-ink-700">
              CargoPrener is for shops and workshops that cannot meet a supplier’s
              minimum on their own. Add only the quantity you need. If the minimum
              is reached before the deadline, the order is bought at the wholesale
              price. If it is not, the order is cancelled and nobody has to buy.
            </p>
            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <Button to="/orders" size="lg">
                Join an order
              </Button>
              <Button to="/orders/new" variant="secondary" size="lg">
                Create an order
              </Button>
            </div>
            <p className="mt-4 text-sm text-ink-700">
              You take on no obligation unless the minimum is reached.
            </p>
          </div>
          <Card className="p-5 sm:p-6">
            <OrderPreview
              supplier={exampleSupplier}
              product={exampleProduct}
              order={exampleOrder}
              currentQty={exampleOrder.collectedQty}
              showScaleNote
            />
            <p className="mt-4 text-sm text-ink-700">
              Example only. Not a live order.
            </p>
          </Card>
        </div>
      </Container>
    </section>
  )
}
