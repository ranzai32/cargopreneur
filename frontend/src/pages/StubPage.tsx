import { useEffect } from 'react'
import { Container } from '../components/Container'
import { PageLayout } from '../components/PageLayout'
import { Section } from '../components/Section'

export function StubPage({ title }: { title: string }) {
  useEffect(() => {
    document.title = `${title} · CargoPrener`
  }, [title])

  return (
    <PageLayout>
      <Section>
        <Container>
          <h1 className="font-serif text-4xl text-ink-950">{title}</h1>
        </Container>
      </Section>
    </PageLayout>
  )
}
