'use client'

import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { Button } from '../../../components/ui/button'
import { Input } from '../../../components/ui/input'
import { api } from '../../../lib/api'

export default function CadastroPage() {
  const router = useRouter()
  const [form, setForm] = useState({ nome: '', email: '', cpf_cnpj: '', senha: '', tipo: 'MEI' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    setForm({ ...form, [e.target.name]: e.target.value })
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await api.post('/api/v1/usuarios', form)

      // Login automático após cadastro — envia JSON body, não Basic Auth
      await api.post('/api/v1/auth/token', {
        username: form.email,
        password: form.senha,
      })
      // Token JWT é definido como cookie httpOnly pelo backend
      router.push('/painel')
    } catch (err: any) {
      setError(err.response?.data?.detail?.[0]?.msg || err.response?.data?.detail || 'Erro ao cadastrar')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full bg-white p-8 rounded-lg shadow-md">
        <h2 className="text-2xl font-bold text-center mb-6">Criar Conta</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input name="nome" placeholder="Nome completo" value={form.nome} onChange={handleChange} required />
          <Input name="email" type="email" placeholder="E-mail" value={form.email} onChange={handleChange} required />
          <Input name="cpf_cnpj" placeholder="CPF (11 dígitos) ou CNPJ (14 dígitos)" value={form.cpf_cnpj} onChange={handleChange} required />
          <Input name="senha" type="password" placeholder="Senha (mín. 8 caracteres)" value={form.senha} onChange={handleChange} required />
          <select
            name="tipo"
            value={form.tipo}
            onChange={handleChange}
            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
          >
            <option value="MEI">MEI</option>
            <option value="EMPRESA">Empresa</option>
          </select>
          {error && <p className="text-red-500 text-sm">{error}</p>}
          <Button type="submit" className="w-full" disabled={loading}>
            {loading ? 'Cadastrando...' : 'Criar Conta'}
          </Button>
        </form>
        <p className="text-center text-sm text-gray-600 mt-4">
          Já tem conta?{' '}
          <Link href="/login" className="text-blue-600 hover:underline">Entrar</Link>
        </p>
      </div>
    </div>
  )
}