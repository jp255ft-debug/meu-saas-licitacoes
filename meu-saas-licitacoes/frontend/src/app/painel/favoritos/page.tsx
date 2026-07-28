"use client"

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { ChevronRight, Eye, Trash2 } from "lucide-react"
import { useRouter } from "next/navigation"
import { Button } from "../../../components/ui/button"
import { Card, CardContent } from "../../../components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../../components/ui/tabs"
import { api } from "../../../lib/api"

const COLUNAS = ["PARA_LER", "LENDO", "ANALISADO", "ARQUIVADO"] as const
const CORES: Record<string, string> = {
    PARA_LER: "bg-gray-100",
    LENDO: "bg-blue-50",
    ANALISADO: "bg-green-50",
    ARQUIVADO: "bg-yellow-50",
}

export default function FavoritosPage() {
    const router = useRouter()
    const queryClient = useQueryClient()

    const { data: favoritos, isLoading } = useQuery({
        queryKey: ["favoritos"],
        queryFn: async () => {
            const response = await api.get("/api/v1/editais/favoritos")
            return response.data
        },
    })

    const removerMutation = useMutation({
        mutationFn: async (editalId: string) => {
            await api.delete(`/api/v1/editais/${editalId}/favoritar`)
        },
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["favoritos"] }),
    })

    const atualizarStatusMutation = useMutation({
        mutationFn: async ({ uuid, status }: { uuid: string; status: string }) => {
            await api.patch(`/api/v1/editais/favoritos/${uuid}`, { status_kanban: status })
        },
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["favoritos"] }),
    })

    const agruparPorColuna = (coluna: string) => {
        return (favoritos || []).filter((f: any) => f.status_kanban === coluna)
    }

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <h1 className="text-3xl font-bold mb-6">Meus Favoritos</h1>

            {isLoading ? (
                <p className="text-muted-foreground">Carregando...</p>
            ) : !favoritos || favoritos.length === 0 ? (
                <Card>
                    <CardContent className="py-12 text-center">
                        <p className="text-muted-foreground mb-4">
                            Você ainda não favoritou nenhum edital.
                        </p>
                        <Button onClick={() => router.push("/painel/editais")}>
                            Explorar Editais
                        </Button>
                    </CardContent>
                </Card>
            ) : (
                <Tabs defaultValue="PARA_LER">
                    <TabsList className="mb-6">
                        {COLUNAS.map((col) => (
                            <TabsTrigger key={col} value={col}>
                                {col.replace("_", " ")} ({agruparPorColuna(col).length})
                            </TabsTrigger>
                        ))}
                    </TabsList>

                    {COLUNAS.map((coluna) => (
                        <TabsContent key={coluna} value={coluna}>
                            <div className="space-y-3">
                                {agruparPorColuna(coluna).map((fav: any) => (
                                    <Card key={fav.uuid} className={CORES[coluna]}>
                                        <CardContent className="p-4">
                                            <div className="flex items-start justify-between">
                                                <div className="flex-1">
                                                    {fav.edital && (
                                                        <>
                                                            <p className="font-medium">
                                                                {fav.edital.objeto?.substring(0, 120)}
                                                            </p>
                                                            <p className="text-sm text-muted-foreground mt-1">
                                                                {fav.edital.orgao_nome} - {fav.edital.uf}
                                                            </p>
                                                            {fav.edital.valor_estimado && (
                                                                <p className="text-sm font-medium text-green-700 mt-1">
                                                                    R$ {fav.edital.valor_estimado.toLocaleString()}
                                                                </p>
                                                            )}
                                                        </>
                                                    )}
                                                    {fav.notas && (
                                                        <p className="text-xs text-muted-foreground mt-2 italic">
                                                            {fav.notas}
                                                        </p>
                                                    )}
                                                </div>

                                                <div className="flex items-center gap-1 ml-4">
                                                    {fav.edital && (
                                                        <Button
                                                            variant="ghost"
                                                            size="sm"
                                                            onClick={() => router.push(`/painel/editais/${fav.edital_mongodb_id}`)}
                                                        >
                                                            <Eye className="h-4 w-4" />
                                                        </Button>
                                                    )}

                                                    {/* Avançar para próxima coluna */}
                                                    {coluna !== "ARQUIVADO" && (
                                                        <Button
                                                            variant="ghost"
                                                            size="sm"
                                                            onClick={() =>
                                                                atualizarStatusMutation.mutate({
                                                                    uuid: fav.uuid,
                                                                    status: COLUNAS[COLUNAS.indexOf(coluna) + 1],
                                                                })
                                                            }
                                                        >
                                                            <ChevronRight className="h-4 w-4" />
                                                        </Button>
                                                    )}

                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        onClick={() => removerMutation.mutate(fav.edital_mongodb_id)}
                                                    >
                                                        <Trash2 className="h-4 w-4 text-red-500" />
                                                    </Button>
                                                </div>
                                            </div>
                                        </CardContent>
                                    </Card>
                                ))}
                            </div>
                        </TabsContent>
                    ))}
                </Tabs>
            )}
        </div>
    )
}