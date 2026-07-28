'use client'

import { useQuery } from '@tanstack/react-query'
import { format } from 'date-fns'
import { Eye } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { Badge } from '../../../components/ui/badge'
import { Button } from '../../../components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '../../../components/ui/card'
import { Input } from '../../../components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table'
import { api } from '../../../lib/api'

export default function PCAPage() {
    const router = useRouter()
    const [ano, setAno] = useState(new Date().getFullYear())
    const [uf, setUf] = useState('')

    const { data: pcaData, isLoading, refetch } = useQuery({
        queryKey: ['pca', ano, uf],
        queryFn: async () => {
            const params = new URLSearchParams({ ano: String(ano) })
            if (uf) params.append('uf', uf)
            const response = await api.get(`/api/v1/pca?${params.toString()}`)
            return response.data
        },
    })

    // Buscar apenas os PCAs compatíveis com o usuário
    const { data: meusPca, isLoading: loadingMeus } = useQuery({
        queryKey: ['meus-pca'],
        queryFn: async () => {
            const response = await api.get('/api/v1/pca/meus-pca')
            return response.data
        },
    })

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-3xl font-bold">Plano de Contratações Anual</h1>
                <Badge className="bg-blue-100 text-blue-800">Plano Pro</Badge>
            </div>

            {/* Seção: PCAs recomendados */}
            <Card className="mb-8 border-blue-200 bg-blue-50">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        🎯 Recomendados para você
                        <Badge className="ml-2" variant="outline">Score</Badge>
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    {loadingMeus ? (
                        <p className="text-muted-foreground">Carregando...</p>
                    ) : meusPca && meusPca.items && meusPca.items.length > 0 ? (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Órgão</TableHead>
                                    <TableHead>UF</TableHead>
                                    <TableHead>Itens</TableHead>
                                    <TableHead>Score</TableHead>
                                    <TableHead>Ações</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {meusPca.items.slice(0, 10).map((pca: any) => (
                                    <TableRow key={pca.id_pca}>
                                        <TableCell className="font-medium">{pca.orgao_entidade?.nome || 'N/A'}</TableCell>
                                        <TableCell>{pca.orgao_entidade?.uf || ''}</TableCell>
                                        <TableCell>{pca.itens?.length || 0}</TableCell>
                                        <TableCell>
                                            <Badge className="bg-green-100 text-green-800">
                                                {pca.score_total || 0}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>
                                            <Button
                                                variant="ghost"
                                                onClick={() => router.push(`/painel/pca/${pca.id_pca}`)}
                                            >
                                                <Eye className="h-4 w-4" />
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    ) : (
                        <p className="text-muted-foreground">
                            Nenhum PCA compatível encontrado. Configure seu Smart Match para receber recomendações.
                        </p>
                    )}
                </CardContent>
            </Card>

            {/* Filtros */}
            <div className="flex gap-4 mb-6 items-end">
                <div>
                    <label className="text-sm text-muted-foreground mb-1 block">Ano</label>
                    <Input
                        type="number"
                        placeholder="Ano"
                        value={ano}
                        onChange={(e) => setAno(Number(e.target.value))}
                        className="w-32"
                    />
                </div>
                <div>
                    <label className="text-sm text-muted-foreground mb-1 block">UF</label>
                    <Input
                        placeholder="Ex: SP"
                        value={uf}
                        onChange={(e) => setUf(e.target.value.toUpperCase())}
                        className="w-32"
                    />
                </div>
                <Button onClick={() => refetch()}>Filtrar</Button>
            </div>

            {/* Lista completa de PCAs */}
            <Card>
                <CardHeader>
                    <CardTitle>Todos os PCAs</CardTitle>
                </CardHeader>
                <CardContent>
                    {isLoading ? (
                        <p className="text-muted-foreground">Carregando...</p>
                    ) : pcaData && pcaData.items && pcaData.items.length > 0 ? (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Órgão</TableHead>
                                    <TableHead>UF</TableHead>
                                    <TableHead>Itens</TableHead>
                                    <TableHead>Publicação</TableHead>
                                    <TableHead>Ações</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {pcaData.items.map((pca: any) => (
                                    <TableRow key={pca._id}>
                                        <TableCell className="font-medium">{pca.orgao_entidade?.nome || 'N/A'}</TableCell>
                                        <TableCell>{pca.orgao_entidade?.uf || ''}</TableCell>
                                        <TableCell>{pca.itens?.length || 0}</TableCell>
                                        <TableCell>
                                            {pca.data_publicacao
                                                ? format(new Date(pca.data_publicacao), 'dd/MM/yyyy')
                                                : 'N/A'}
                                        </TableCell>
                                        <TableCell>
                                            <Button
                                                variant="ghost"
                                                onClick={() => router.push(`/painel/pca/${pca.id_pca}`)}
                                            >
                                                <Eye className="h-4 w-4" />
                                            </Button>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    ) : (
                        <p className="text-muted-foreground">Nenhum PCA encontrado para os filtros selecionados.</p>
                    )}
                    {pcaData && (
                        <p className="text-sm text-muted-foreground mt-4">
                            Total: {pcaData.total} PCAs • Página {pcaData.pagina} de {pcaData.total_paginas}
                        </p>
                    )}
                </CardContent>
            </Card>
        </div>
    )
}