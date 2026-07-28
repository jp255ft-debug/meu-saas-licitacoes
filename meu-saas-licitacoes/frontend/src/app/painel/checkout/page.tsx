"use client"

import { useMutation } from "@tanstack/react-query"
import { Check, Sparkles } from "lucide-react"
import { useRouter } from "next/navigation"
import { useState } from "react"
import { Badge } from "../../../components/ui/badge"
import { Button } from "../../../components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "../../../components/ui/card"
import { api } from "../../../lib/api"

const PLANOS = [
    {
        id: "starter",
        nome: "Starter",
        preco_mensal: 97,
        preco_anual: 970,
        recursos: [
            "3 palavras-chave para busca",
            "Alertas diários por e-mail",
            "Lista de editais públicos",
            "Suporte por e-mail",
        ],
        destaque: false,
    },
    {
        id: "pro",
        nome: "Pro",
        preco_mensal: 197,
        preco_anual: 1970,
        recursos: [
            "Palavras-chave ilimitadas",
            "Alertas WhatsApp + E-mail + In-app",
            "Resumo com IA (5 tópicos)",
            "Kanban de favoritos",
            "Plano de Contratações Anual (PCA)",
            "Suporte prioritário",
        ],
        destaque: true,
    },
]

export default function CheckoutPage() {
    const router = useRouter()
    const [ciclo, setCiclo] = useState<"mensal" | "anual">("mensal")

    const checkoutMutation = useMutation({
        mutationFn: async (plano: string) => {
            const response = await api.post("/api/v1/assinaturas/criar-checkout", {
                plano,
                ciclo,
                success_url: `${window.location.origin}/painel?checkout=sucesso`,
                cancel_url: `${window.location.origin}/painel/checkout`,
            })
            return response.data
        },
        onSuccess: (data) => {
            if (data.checkout_url) {
                window.location.href = data.checkout_url
            }
        },
    })

    return (
        <div className="p-6 max-w-5xl mx-auto">
            <div className="text-center mb-8">
                <h1 className="text-3xl font-bold mb-2">Escolha seu Plano</h1>
                <p className="text-muted-foreground">
                    Comece grátis por 7 dias. Cancele quando quiser.
                </p>
            </div>

            {/* Toggle ciclo */}
            <div className="flex justify-center mb-8">
                <div className="inline-flex bg-gray-100 rounded-lg p-1">
                    <button
                        className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${ciclo === "mensal" ? "bg-white shadow-sm" : "hover:bg-gray-200"
                            }`}
                        onClick={() => setCiclo("mensal")}
                    >
                        Mensal
                    </button>
                    <button
                        className={`px-4 py-2 rounded-md text-sm font-medium transition-colors flex items-center gap-1 ${ciclo === "anual" ? "bg-white shadow-sm" : "hover:bg-gray-200"
                            }`}
                        onClick={() => setCiclo("anual")}
                    >
                        Anual
                        <Badge className="bg-green-100 text-green-800 text-xs">-10%</Badge>
                    </button>
                </div>
            </div>

            {/* Cards de planos */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {PLANOS.map((plano) => (
                    <Card
                        key={plano.id}
                        className={`relative ${plano.destaque ? "border-blue-400 shadow-lg ring-2 ring-blue-400" : ""}`}
                    >
                        {plano.destaque && (
                            <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                                <Badge className="bg-blue-600 text-white px-4 py-1">
                                    <Sparkles className="h-3 w-3 mr-1" />
                                    Mais Popular
                                </Badge>
                            </div>
                        )}

                        <CardHeader className={plano.destaque ? "pt-8" : ""}>
                            <CardTitle className="text-2xl">{plano.nome}</CardTitle>
                            <div className="mt-2">
                                <span className="text-4xl font-bold">
                                    R$ {ciclo === "mensal" ? plano.preco_mensal : (plano.preco_anual / 12).toFixed(0)}
                                </span>
                                <span className="text-muted-foreground ml-1">/mês</span>
                            </div>
                            {ciclo === "anual" && (
                                <p className="text-sm text-green-600 mt-1">
                                    R$ {plano.preco_anual}/ano (economia de R$ {(plano.preco_mensal * 12 - plano.preco_anual).toFixed(0)})
                                </p>
                            )}
                        </CardHeader>

                        <CardContent>
                            <ul className="space-y-3 mb-6">
                                {plano.recursos.map((recurso) => (
                                    <li key={recurso} className="flex items-start gap-2">
                                        <Check className="h-5 w-5 text-green-500 shrink-0 mt-0.5" />
                                        <span className="text-sm">{recurso}</span>
                                    </li>
                                ))}
                            </ul>

                            <Button
                                className="w-full"
                                variant={plano.destaque ? "default" : "outline"}
                                disabled={checkoutMutation.isPending}
                                onClick={() => checkoutMutation.mutate(plano.id)}
                            >
                                {checkoutMutation.isPending ? "Redirecionando..." : "Assinar Agora"}
                            </Button>

                            <p className="text-xs text-muted-foreground text-center mt-3">
                                Teste grátis por 7 dias. Cancele antes de pagar.
                            </p>
                        </CardContent>
                    </Card>
                ))}
            </div>
        </div>
    )
}