import { useEffect } from 'react'
import { Benefits } from '../components/landing/Benefits'
import { FinalCta } from '../components/landing/FinalCta'
import { Hero } from '../components/landing/Hero'
import { HowItWorks } from '../components/landing/HowItWorks'
import { OrderExample } from '../components/landing/OrderExample'
import { ProblemSolution } from '../components/landing/ProblemSolution'
import { Trust } from '../components/landing/Trust'
import { PageLayout } from '../components/PageLayout'

export function LandingPage() {
  useEffect(() => {
    document.title = 'CargoPrener — Pool orders and reach wholesale prices'
  }, [])

  return (
    <PageLayout>
      <Hero />
      <ProblemSolution />
      <HowItWorks />
      <OrderExample />
      <Benefits />
      <Trust />
      <FinalCta />
    </PageLayout>
  )
}
