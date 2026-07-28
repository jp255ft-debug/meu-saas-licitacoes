"use client"

import { useQuery } from "@tanstack/react-query"
import { format } from "date-fns"
import { Eye } from "lucide-react"
import { useRouter } from "next/navigation"
import { useState } from "react"
import { Badge } from "../../../components/ui/badge"
import { Button } from "../../../components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "../../../components/ui/card"
import { Input } from "../../../components/ui/input"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../../../components/ui/table"
import { api } from "../../../lib/api"

export default function EditaisPage() {
    const router = useRouter()
    const [palavraChave, setPalavraChave] = useState("")
    const [uf, setUf] = useState("")
    const [pagina, setPagina] = useState(1)

    const { data, isLoading } = useQuery({
        queryKey: ["editais", palavraChave, uf, pagina],
        queryFn: async () => {
            const params: Record<string, string> = {
                pagina: String(pagina),
                tamanho: "20",
            }
            if (palavraChave) params.palavra_chave = palavraChave
            if (uf) params.uf = uf
            const response = await api.get("/api/v1/editais", { params })
            return response.data
        },
    })

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <h1 className="text-3xl font-bold mb-6">Editais Públicos</h1>

            {/* Filtros */}
            <Card className="mb-6">
                <CardContent className="pt-6">
                    <div className="flex gap-4 items-end">
                        <div className="flex-1">
                            <label className="text-sm text-muted-foreground mb-1 block">Palavra-chave</label>
                            <Input
                                placeholder="Buscar no objeto ou órgão..."
                                value={palavraChave}
                                onChange={(e) => { setPalavraChave(e.target.value); setPagina(1) }}
                            />
                        </div>
                        <div>
                            <label className="text-sm text-muted-foreground mb-1 block">UF</label>
                            <Input
                                placeholder="Ex: SP"
                                value={uf}
                                onChange={(e) => { setUf(e.target.value.toUpperCase()); setPagina(1) }}
                                className="w-24"
                            />
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Lista */}
            <Card>
                <CardHeader>
                    <CardTitle>
                        Resultados
                        {data && <span className="text-sm font-normal text-muted-foreground ml-2">({data.paginacao?.total || 0} encontrados)</span>}
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <p className="text-muted-foreground">Carregando...</p>
                    ) : data && data.data && data.data.length > 0 ? (
                        <>
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Objeto</TableHead>
                                        <TableHead>Órgão</TableHead>
                                        <TableHead>Valor</TableHead>
                                        <TableHead>Prazo</TableHead>
                                        <TableHead></TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {data.data.map((edital: any) => (
                                        <TableRow key={edital.id}>
                                            <TableCell className="max-w-md">
                                                <p className="font-medium truncate">{edital.objeto?.substring(0, 80)}...</p>
                                                <div className="flex gap-1 mt-1">
                                                    {edital.categorias?.slice(0, 2).map((cat: string) => (
                                                        <Badge key={cat} variant="outline" className="text-xs">{cat}</Badge>
                                                    ))}
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                <p className="text-sm">{edital.orgao_nome?.substring(0, 40)}</p>
                                                <p className="text-xs text-muted-foreground">{edital.uf}</p>
                                            </TableCell>
                                            <TableCell>
                                                <p className="font-medium">R$ {edital.valor_estimado?.toLocaleString() || "N/A"}</p>
                                            </TableCell>
                                            <TableCell className="text-sm">
                                                {edital.data_abertura
                                                    ? format(new Date(edital.data_abertura), "dd/MM/yyyy")
                                                    : "N/A"}
                                            </TableCell>
                                            <TableCell>
                                                <Button variant="ghost" onClick={() => router.push(`/painel/editais/${edital.id}`)}>
                                                    <Eye className="h-4 w-4" />
                                                </Button>
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>

                            {/* Paginação */}
                            {data.paginacao && data.paginacao.total_paginas > 1 && (
                                <div className="flex items-center justify-between mt-4">
                                    <p className="text-sm text-muted-foreground">
                                        Página {pagina} de {data.paginacao.total_paginas}
                                    </p>
                                    <div className="flex gap-2">
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            disabled={pagina <= 1}
                                            onClick={() => setPagina((p) => Math.max(1, p - 1))}
                                        >
                                            Anterior
                                        </Button>
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            disabled={pagina >= data.paginacao.total_paginas}
                                            onClick={() => setPagina((p) => p + 1)}
                                        >
                                            Próxima
                                        </Button>
                                    </div>
                                </div>
                            )}
                        </>
                    ) : (
                        <p className="text-muted-foreground">Nenhum edital encontrado.</p>
                    )}
                </CardContent>
            </Card>
        </div>
    )
}