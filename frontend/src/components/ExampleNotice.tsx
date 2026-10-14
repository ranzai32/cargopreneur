import { exampleMode } from '../lib/api'
import { Container } from './Container'

export function ExampleNotice() {
  if (!exampleMode) return null
  return (
    <div className="border-b border-warning-soft bg-warning-soft">
      <Container className="py-2 text-sm text-warning">
        Example data in this browser. Sign-in and new orders stay on this device until the API is connected.
      </Container>
    </div>
  )
}
