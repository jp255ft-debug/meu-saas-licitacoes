import Link from 'next/link'
import { Button } from '../components/ui/button'

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-blue-50 to-white">
      <header className="container mx-auto py-6 flex justify-between items-center px-4">
        <h1 className="text-2xl font-bold text-blue-600">Licitação SaaS</h1>
        <div className="space-x-4">
          <Link href="/login">
            <Button variant="outline">Entrar</Button>
          </Link>
          <Link href="/cadastro">
            <Button>Começar Grátis</Button>
          </Link>
        </div>
      </header>

      <main className="container mx-auto px-4 py-16">
        <section className="text-center max-w-3xl mx-auto">
          <h2 className="text-4xl md:text-5xl font-bold text-gray-900 mb-6">
            Encontre oportunidades públicas com <span className="text-blue-600">inteligência</span>
          </h2>
          <p className="text-lg text-gray-600 mb-8">
            Resumimos editais de 80 páginas em 5 tópicos. Alertas de vencimento de certidões.
            Smart Match para o seu perfil de empresa.
          </p>
          <Link href="/cadastro">
            <Button size="lg" className="text-lg px-8">
              Começar agora - Grátis por 7 dias
            </Button>
          </Link>
        </section>

        <section className="mt-20 grid md:grid-cols-3 gap-8 max-w-5xl mx-auto">
          <div className="bg-white p-6 rounded-lg shadow-sm border">
            <h3 className="text-xl font-semibold mb-2">Smart Match</h3>
            <p className="text-gray-600">Filtros por palavra-chave e porte da empresa. Só o que interessa para você.</p>
          </div>
          <div className="bg-white p-6 rounded-lg shadow-sm border">
            <h3 className="text-xl font-semibold mb-2">Resumo com IA</h3>
            <p className="text-gray-600">Leia o objeto, valor, exclusividade ME/EPP e prazos em segundos.</p>
          </div>
          <div className="bg-white p-6 rounded-lg shadow-sm border">
            <h3 className="text-xl font-semibold mb-2">Radar de Certidões</h3>
            <p className="text-gray-600">Alertas 7 dias antes do vencimento. Nunca mais perca uma licitação por documentação.</p>
          </div>
        </section>
      </main>
    </div>
  )
}