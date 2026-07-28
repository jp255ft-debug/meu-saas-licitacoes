"use client"

import { useQuery } from "@tanstack/react-query"
import { format } from "date-fns"
import {
    CalendarCheck,
    FileText,
    Star,
    TrendingUp
} from "lucide-react"
import { useRouter } from "next/navigation"
import { Badge } from "../../components/ui/badge"
import { Button } from "../../components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "../../components/ui/card"
import { api } from "../../lib/api"

export default function DashboardPage() {
    const router = useRouter()

    const { data: contratacoes } = useQuery({
        queryKey: ["contratacoes-stats"],
        queryFn: async () => {
            const response = await api.get("/api/v1/contratacoes/estatisticas")
            return response.data
        },
    })

    const { data: pcaStats } = useQuery({
        queryKey: ["pca-stats"],
        queryFn: async () => {
            const response = await api.get("/api/v1/pca/estatisticas")
            return response.data
        },
    })

    const { data: ultimas } = useQuery({
        queryKey: ["ultimas-contratacoes"],
        queryFn: async () => {
            const response = await api.get("/api/v1/contratacoes", {
                params: { pagina: 1, tamanho: 5 },
            })
            return response.data
        },
    })

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <h1 className="text-3xl font-bold mb-6">Dashboard</h1>

            {/* Cards de estatísticas */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium">Contratações Abertas</CardTitle>
                        <FileText className="h-4 w-4 text-blue-600" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{contratacoes?.total_abertas || 0}</div>
                        <p className="text-xs text-muted-foreground">Com propostas abertas</p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium">Exclusivas ME/EPP</CardTitle>
                        <Star className="h-4 w-4 text-yellow-600" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{contratacoes?.exclusivas_me_epp || 0}</div>
                        <p className="text-xs text-muted-foreground">Prioridade para pequenas empresas</p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium">PCAs Disponíveis</CardTitle>
                        <CalendarCheck className="h-4 w-4 text-green-600" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{pcaStats?.total_pcas_ano_atual || 0}</div>
                        <p className="text-xs text-muted-foreground">Planejamento anual</p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium">Valor Total</CardTitle>
                        <TrendingUp className="h-4 w-4 text-green-600" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            R$ {contratacoes?.valor_total_estimado?.toLocaleString() || "0"}
                        </div>
                        <p className="text-xs text-muted-foreground">Em contratações abertas</p>
                    </CardContent>
                </Card>
            </div>

            {/* Seção: Últimas oportunidades */}
            <Card className="mb-8">
                <CardHeader className="flex flex-row items-center justify-between">
                    <CardTitle>Últimas Oportunidades</CardTitle>
                    <Button variant="outline" size="sm" onClick={() => router.push("/painel/editais")}>
                        Ver todas
                    </Button>
                </CardHeader>
                <CardContent>
                    {ultimas && ultimas.data && ultimas.data.length > 0 ? (
                        <div className="space-y-4">
                            {ultimas.data.slice(0, 5).map((item: any) => (
                                <div
                                    key={item.id}
                                    className="flex items-start justify-between p-4 border rounded-lg hover:bg-gray-50 cursor-pointer"
                                    onClick={() => router.push(`/painel/editais/${item.id}`)}
                                >
                                    <div className="flex-1">
                                        <p className="font-medium">{item.objeto_compra?.substring(0, 100)}</p>
                                        <p className="text-sm text-muted-foreground mt-1">
                                            {item.orgao_entidade?.nome || "N/A"} - {item.orgao_entidade?.uf || ""}
                                        </p>
                                        <div className="flex gap-2 mt-2">
                                            {item.exclusividade_me_epp && (
                                                <Badge className="bg-yellow-100 text-yellow-800 text-xs">ME/EPP</Badge>
                                            )}
                                            {item.modalidade_nome && (
                                                <Badge variant="outline" className="text-xs">{item.modalidade_nome}</Badge>
                                            )}
                                        </div>
                                    </div>
                                    <div className="text-right ml-4">
                                        <p className="font-bold text-green-700">
                                            R$ {item.valor_total_estimado?.toLocaleString() || "N/A"}
                                        </p>
                                        <p className="text-xs text-muted-foreground mt-1">
                                            {item.data_encerramento_proposta
                                                ? format(new Date(item.data_encerramento_proposta), "dd/MM")
                                                : "N/A"}
                                        </p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <p className="text-muted-foreground">Nenhuma contratação encontrada.</p>
                    )}
                </CardContent>
            </Card>

            {/* Seção: Acesso rápido */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <Card className="bg-blue-50 border-blue-200">
                    <CardHeader>
                        <CardTitle className="text-lg">🎯 Smart Match</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-sm text-muted-foreground mb-4">
                            Configure seu perfil para receber recomendações personalizadas.
                        </p>
                        <Button onClick={() => router.push("/painel/configuracoes")}>
                            Configurar
                        </Button>
                    </CardContent>
                </Card>

                <Card className="bg-green-50 border-green-200">
                    <CardHeader>
                        <CardTitle className="text-lg">📋 PCA</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-sm text-muted-foreground mb-4">
                            Anteveja compras públicas com o Plano de Contratações Anual.
                        </p>
                        <Button onClick={() => router.push("/painel/pca")}>
                            Ver PCAs
                        </Button>
                    </CardContent>
                </Card>

                <Card className="bg-purple-50 border-purple-200">
                    <CardHeader>
                        <CardTitle className="text-lg">⭐ Pro</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-sm text-muted-foreground mb-4">
                            Desbloqueie recursos avançados com o plano Pro.
                        </p>
                        <Button onClick={() => router.push("/painel/checkout")}>
                            Ver Planos
                        </Button>
                    </CardContent>
                </Card>
            </div>
        </div>
    )
}