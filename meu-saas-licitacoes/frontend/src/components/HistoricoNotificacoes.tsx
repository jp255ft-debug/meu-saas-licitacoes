"use client"

import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { api } from "../lib/api"
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../components/ui/table"
import { Badge } from "../components/ui/badge"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../components/ui/select"
import { Button } from "../components/ui/button"

const statusColor: Record<string, string> = {
  pendente: "bg-yellow-100 text-yellow-800 border-yellow-200",
  enviado: "bg-green-100 text-green-800 border-green-200",
  falha: "bg-red-100 text-red-800 border-red-200",
  lido: "bg-blue-100 text-blue-800 border-blue-200",
}

const statusLabel: Record<string, string> = {
  pendente: "Pendente",
  enviado: "Enviado",
  falha: "Falha",
  lido: "Lido",
}

interface Notificacao {
  uuid: string
  tipo: string
  status: string
  contratacao_id: string
  mensagem: string
  data_envio: string | null
  data_leitura: string | null
  erro: string | null
  tentativas: number
  criado_em: string
}

interface NotificacoesResponse {
  items: Notificacao[]
  total: number
  pagina: number
  tamanho: number
  total_paginas: number
}

export function HistoricoNotificacoes() {
  const [tipo, setTipo] = useState("")
  const [status, setStatus] = useState("")
  const [pagina, setPagina] = useState(1)

  const { data, isLoading } = useQuery<NotificacoesResponse>({
    queryKey: ["notificacoes", tipo, status, pagina],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (tipo) params.append("tipo", tipo)
      if (status) params.append("status", status)
      params.append("pagina", String(pagina))
      params.append("tamanho", "20")
      const response = await api.get(`/api/v1/usuarios/me/notificacoes?${params.toString()}`)
      return response.data
    },
  })

  if (isLoading) return <div className="p-6 text-center text-gray-500">Carregando...</div>

  return (
    <Card>
      <CardHeader>
        <CardTitle>Histórico de Notificações</CardTitle>
        <div className="flex gap-4 mt-2">
          <Select value={tipo} onValueChange={(v) => { setTipo(v); setPagina(1) }}>
            <SelectTrigger className="w-[150px]">
              <SelectValue placeholder="Todos os tipos" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="">Todos</SelectItem>
              <SelectItem value="whatsapp">WhatsApp</SelectItem>
              <SelectItem value="email">E-mail</SelectItem>
              <SelectItem value="in_app">In-App</SelectItem>
            </SelectContent>
          </Select>
          <Select value={status} onValueChange={(v) => { setStatus(v); setPagina(1) }}>
            <SelectTrigger className="w-[150px]">
              <SelectValue placeholder="Todos os status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="">Todos</SelectItem>
              <SelectItem value="pendente">Pendente</SelectItem>
              <SelectItem value="enviado">Enviado</SelectItem>
              <SelectItem value="falha">Falha</SelectItem>
              <SelectItem value="lido">Lido</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardHeader>
      <CardContent>
        {data && data.items.length > 0 ? (
          <>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Data</TableHead>
                  <TableHead>Tipo</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Mensagem</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.items.map((notif: Notificacao) => (
                  <TableRow key={notif.uuid}>
                    <TableCell className="text-sm">
                      {new Date(notif.criado_em).toLocaleDateString("pt-BR", {
                        day: "2-digit",
                        month: "2-digit",
                        year: "numeric",
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </TableCell>
                    <TableCell className="capitalize text-sm">{notif.tipo === "in_app" ? "In-App" : notif.tipo}</TableCell>
                    <TableCell>
                      <Badge className={statusColor[notif.status] || ""}>
                        {statusLabel[notif.status] || notif.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="max-w-xs truncate text-sm">{notif.mensagem}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            {data.total_paginas > 1 && (
              <div className="flex justify-center gap-2 mt-4">
                {Array.from({ length: data.total_paginas }, (_, i) => i + 1).map((p) => (
                  <Button
                    key={p}
                    variant={p === pagina ? "default" : "outline"}
                    size="sm"
                    onClick={() => setPagina(p)}
                  >
                    {p}
                  </Button>
                ))}
              </div>
            )}
          </>
        ) : (
          <p className="text-gray-500 text-center py-8">Nenhuma notificação encontrada.</p>
        )}
      </CardContent>
    </Card>
  )
}