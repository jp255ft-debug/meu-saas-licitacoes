"use client"

import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../../components/ui/tabs"
import { ConfiguracoesAlerta } from "../../../components/ConfiguracoesAlerta"
import { ConfiguracoesSmartMatch } from "../../../components/ConfiguracoesSmartMatch"
import { HistoricoNotificacoes } from "../../../components/HistoricoNotificacoes"

export default function ConfiguracoesPage() {
  return (
    <div className="p-6 max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold mb-6">Configurações</h1>
      <Tabs defaultValue="alertas" className="space-y-4">
        <TabsList>
          <TabsTrigger value="alertas">Alertas</TabsTrigger>
          <TabsTrigger value="smart-match">Smart Match</TabsTrigger>
          <TabsTrigger value="historico">Histórico</TabsTrigger>
        </TabsList>
        <TabsContent value="alertas">
          <ConfiguracoesAlerta />
        </TabsContent>
        <TabsContent value="smart-match">
          <ConfiguracoesSmartMatch />
        </TabsContent>
        <TabsContent value="historico">
          <HistoricoNotificacoes />
        </TabsContent>
      </Tabs>
    </div>
  )
}