"use client"

import { useState, useEffect } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { api } from "../lib/api"
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card"
import { Switch } from "../components/ui/switch"
import { Input } from "../components/ui/input"
import { Button } from "../components/ui/button"
import { Label } from "../components/ui/label"
import { toast } from "sonner"

export function ConfiguracoesAlerta() {
  const queryClient = useQueryClient()
  const [telefone, setTelefone] = useState("")
  const [alertas, setAlertas] = useState({
    alertas_whatsapp: true,
    alertas_email: true,
    alertas_in_app: true,
  })

  const { data: perfil, isLoading } = useQuery({
    queryKey: ["perfil-usuario"],
    queryFn: async () => {
      const response = await api.get("/api/v1/usuarios/me")
      return response.data
    },
  })

  useEffect(() => {
    if (perfil) {
      setAlertas({
        alertas_whatsapp: perfil.alertas_whatsapp ?? true,
        alertas_email: perfil.alertas_email ?? true,
        alertas_in_app: perfil.alertas_in_app ?? true,
      })
      setTelefone(perfil.telefone_whatsapp || "")
    }
  }, [perfil])

  const mutation = useMutation({
    mutationFn: async (dados: Record<string, unknown>) => {
      const response = await api.patch("/api/v1/usuarios/me/preferencias-alertas", dados)
      return response.data
    },
    onSuccess: () => {
      toast.success("Preferências atualizadas com sucesso!")
      queryClient.invalidateQueries({ queryKey: ["perfil-usuario"] })
    },
    onError: (error: { response?: { data?: { detail?: string } } }) => {
      toast.error(error.response?.data?.detail || "Erro ao atualizar preferências")
    },
  })

  const handleSalvar = () => {
    mutation.mutate({
      ...alertas,
      telefone_whatsapp: telefone,
    })
  }

  if (isLoading) return <div className="p-6 text-center text-gray-500">Carregando...</div>

  return (
    <Card>
      <CardHeader>
        <CardTitle>Canais de Alerta</CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <Label htmlFor="whatsapp" className="font-medium">WhatsApp</Label>
            <p className="text-sm text-gray-500">Receba alertas no seu WhatsApp</p>
          </div>
          <Switch
            id="whatsapp"
            checked={alertas.alertas_whatsapp}
            onCheckedChange={(checked: boolean) =>
              setAlertas((prev) => ({ ...prev, alertas_whatsapp: checked }))
            }
          />
        </div>

        <div className="flex items-center justify-between">
          <div>
            <Label htmlFor="email" className="font-medium">E-mail</Label>
            <p className="text-sm text-gray-500">Receba alertas no seu e-mail</p>
          </div>
          <Switch
            id="email"
            checked={alertas.alertas_email}
            onCheckedChange={(checked: boolean) =>
              setAlertas((prev) => ({ ...prev, alertas_email: checked }))
            }
          />
        </div>

        <div className="flex items-center justify-between">
          <div>
            <Label htmlFor="in_app" className="font-medium">In-App</Label>
            <p className="text-sm text-gray-500">Receba notificações dentro da plataforma</p>
          </div>
          <Switch
            id="in_app"
            checked={alertas.alertas_in_app}
            onCheckedChange={(checked: boolean) =>
              setAlertas((prev) => ({ ...prev, alertas_in_app: checked }))
            }
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="telefone">Telefone WhatsApp</Label>
          <Input
            id="telefone"
            placeholder="(11) 99999-9999"
            value={telefone}
            onChange={(e) => setTelefone(e.target.value)}
            disabled={!alertas.alertas_whatsapp}
          />
          <p className="text-sm text-gray-500">Número com DDD, sem espaços ou caracteres especiais</p>
        </div>

        <Button onClick={handleSalvar} disabled={mutation.isPending}>
          {mutation.isPending ? "Salvando..." : "Salvar Preferências"}
        </Button>
      </CardContent>
    </Card>
  )
}