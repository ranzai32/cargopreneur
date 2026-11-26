export type DeliveryChoice = {
  id: 'pickup' | 'freight' | 'express'
  name: string
  detail: string
  price: number
}

export function deliveryChoices(quantity: number): DeliveryChoice[] {
  const qty = Math.max(1, Math.round(quantity))
  return [
    {
      id: 'pickup',
      name: 'Warehouse pickup',
      detail: 'Collect from the supplier. No freight charge.',
      price: 0,
    },
    {
      id: 'freight',
      name: 'Standard freight',
      detail: '5 to 8 business days to the address you choose.',
      price: cents(2800 + 45 * qty),
    },
    {
      id: 'express',
      name: 'Express freight',
      detail: '2 business days to the address you choose.',
      price: cents(6400 + 90 * qty),
    },
  ]
}

export function deliveryChoice(id: string | null | undefined, quantity: number) {
  return deliveryChoices(quantity).find((item) => item.id === id) ?? null
}

function cents(value: number) {
  return value / 100
}
