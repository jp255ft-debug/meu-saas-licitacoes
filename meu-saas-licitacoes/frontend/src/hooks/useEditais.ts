'use client'

import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'

interface Edital {
  id: string
  pncp_id: string
  orgao_nome: string
  uf: string
  modalidade: string
  objeto: string
  valor_estimado?: number
  data_publicacao?: string
  data_abertura?: string
  resumo?: string
  categorias: string[]
  score_relevancia?: number
}

interface EditaisResponse {
  data: Edital[]
  paginacao: {
    pagina: number
    tamanho: number
    total: number
    total_paginas: number
  }
}

export function useEditais(params?: Record<string, string>) {
  return useQuery<EditaisResponse>({
    queryKey: ['editais', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams(params || {})
      const response = await api.get(`/api/v1/editais?${searchParams.toString()}`)
      return response.data
    },
  })
}

export function useEditaisRelevantes(params?: Record<string, string>) {
  return useQuery<EditaisResponse>({
    queryKey: ['editais-relevantes', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams(params || {})
      const response = await api.get(`/api/v1/editais/relevantes?${searchParams.toString()}`)
      return response.data
    },
  })
}

export function useEdital(id: string) {
  return useQuery<Edital>({
    queryKey: ['edital', id],
    queryFn: async () => {
      const response = await api.get(`/api/v1/editais/${id}`)
      return response.data
    },
    enabled: !!id,
  })
}