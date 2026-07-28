import axios from 'axios'

export const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000',
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,  // envia cookies httpOnly automaticamente
})

// Interceptor para redirecionar ao expirar token (401)
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401 && typeof window !== 'undefined') {
      // Token inválido ou expirado → redirecionar para login
      // Não precisa limpar localStorage — o token está no cookie httpOnly
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)