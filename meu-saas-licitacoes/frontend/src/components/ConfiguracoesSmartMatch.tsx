"use client"

import { useState, useEffect } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { api } from "../lib/api"
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card"
import { Input } from "../components/ui/input"
import { Button } from "../components/ui/button"
import { Label } from "../components/ui/label"
import { Badge } from "../components/ui/badge"
import { X } from "lucide-react"
import { toast } from "sonner"

export function ConfiguracoesSmartMatch() {
  const queryClient = useQueryClient()
  const [cnaes, setCnaes] = useState<string[]>([])
  const [novoCnae, setNovoCnae] = useState("")
  const [ufs, setUfs] = useState<string[]>([])
  const [novaUf, setNovaUf] = useState("")
  const [palavras, setPalavras] = useState<string[]>([])
  const [novaPalavra, setNovaPalavra] = useState("")

  const { data: perfil, isLoading } = useQuery({
    queryKey: ["perfil-usuario"],
    queryFn: async () => {
      const response = await api.get("/api/v1/usuarios/me")
      return response.data
    },
  })

  useEffect(() => {
    if (perfil?.configuracoes_match) {
      setCnaes(perfil.configuracoes_match.cnaes || [])
      setUfs(perfil.configuracoes_match.ufs || [])
      setPalavras(perfil.configuracoes_match.palavras_chave || [])
    }
  }, [perfil])

  const mutation = useMutation({
    mutationFn: async (dados: Record<string, unknown>) => {
      const response = await api.patch("/api/v1/usuarios/me/preferencias-alertas", {
        configuracoes_match: dados,
      })
      return response.data
    },
    onSuccess: () => {
      toast.success("Smart Match atualizado!")
      queryClient.invalidateQueries({ queryKey: ["perfil-usuario"] })
    },
    onError: (error: { response?: { data?: { detail?: string } } }) => {
      toast.error(error.response?.data?.detail || "Erro ao atualizar Smart Match")
    },
  })

  const handleSalvar = () => {
    mutation.mutate({
      cnaes,
      ufs,
      palavras_chave: palavras,
    })
  }

  const adicionarItem = (lista: string[], setLista: (l: string[]) => void, valor: string) => {
    const trimmed = valor.trim()
    if (trimmed && !lista.includes(trimmed)) {
      setLista([...lista, trimmed])
    }
  }

  const removerItem = (lista: string[], setLista: (l: string[]) => void, valor: string) => {
    setLista(lista.filter((item) => item !== valor))
  }

  if (isLoading) return <div className="p-6 text-center text-gray-500">Carregando...</div>

  return (
    <Card>
      <CardHeader>
        <CardTitle>Smart Match</CardTitle>
        <p className="text-sm text-gray-500">
          Defina seu perfil para receber alertas personalizados
        </p>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* CNAEs */}
        <div className="space-y-2">
          <Label>CNAEs</Label>
          <div className="flex gap-2">
            <Input
              placeholder="Ex: 9511-8/00"
              value={novoCnae}
              onChange={(e) => setNovoCnae(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  adicionarItem(cnaes, setCnaes, novoCnae)
                  setNovoCnae("")
                }
              }}
            />
            <Button
              onClick={() => {
                adicionarItem(cnaes, setCnaes, novoCnae)
                setNovoCnae("")
              }}
            >
              Adicionar
            </Button>
          </div>
          <div className="flex flex-wrap gap-2 mt-2">
            {cnaes.map((cnae) => (
              <Badge key={cnae} className="flex items-center gap-1">
                {cnae}
                <X
                  className="h-3 w-3 cursor-pointer"
                  onClick={() => removerItem(cnaes, setCnaes, cnae)}
                />
              </Badge>
            ))}
          </div>
        </div>

        {/* UFs */}
        <div className="space-y-2">
          <Label>UF de Interesse</Label>
          <div className="flex gap-2">
            <Input
              placeholder="Ex: SP"
              maxLength={2}
              value={novaUf}
              onChange={(e) => setNovaUf(e.target.value.toUpperCase())}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  adicionarItem(ufs, setUfs, novaUf)
                  setNovaUf("")
                }
              }}
            />
            <Button
              onClick={() => {
                adicionarItem(ufs, setUfs, novaUf)
                setNovaUf("")
              }}
            >
              Adicionar
            </Button>
          </div>
          <div className="flex flex-wrap gap-2 mt-2">
            {ufs.map((uf) => (
              <Badge key={uf} className="flex items-center gap-1">
                {uf}
                <X
                  className="h-3 w-3 cursor-pointer"
                  onClick={() => removerItem(ufs, setUfs, uf)}
                />
              </Badge>
            ))}
          </div>
        </div>

        {/* Palavras-chave */}
        <div className="space-y-2">
          <Label>Palavras-chave</Label>
          <div className="flex gap-2">
            <Input
              placeholder="Ex: material de expediente"
              value={novaPalavra}
              onChange={(e) => setNovaPalavra(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  adicionarItem(palavras, setPalavras, novaPalavra)
                  setNovaPalavra("")
                }
              }}
            />
            <Button
              onClick={() => {
                adicionarItem(palavras, setPalavras, novaPalavra)
                setNovaPalavra("")
              }}
            >
              Adicionar
            </Button>
          </div>
          <div className="flex flex-wrap gap-2 mt-2">
            {palavras.map((palavra) => (
              <Badge key={palavra} className="flex items-center gap-1">
                {palavra}
                <X
                  className="h-3 w-3 cursor-pointer"
                  onClick={() => removerItem(palavras, setPalavras, palavra)}
                />
              </Badge>
            ))}
          </div>
        </div>

        <Button onClick={handleSalvar} disabled={mutation.isPending}>
          {mutation.isPending ? "Salvando..." : "Salvar Smart Match"}
        </Button>
      </CardContent>
    </Card>
  )
}