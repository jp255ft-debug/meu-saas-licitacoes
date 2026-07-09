import { ApiProperty } from '@nestjs/swagger';

export class ErrorDetailDto {
  @ApiProperty({ example: 'amount_cents', description: 'Campo que gerou o erro' })
  field?: string;

  @ApiProperty({ example: 'Valor deve ser maior que 0', description: 'Descrição do problema' })
  issue?: string;
}

export class ErrorResponseDto {
  @ApiProperty({
    example: {
      code: 'INVALID_INPUT',
      message: 'Dados inválidos',
      details: [{ field: 'amount_cents', issue: 'Valor deve ser maior que 0' }],
    },
  })
  error!: {
    code: string;
    message: string;
    details?: ErrorDetailDto[];
  };
}

export const ERROR_EXAMPLES = {
  badRequest: {
    error: {
      code: 'INVALID_INPUT',
      message: 'Dados inválidos',
      details: [{ field: 'amount_cents', issue: 'Valor deve ser maior que 0' }],
    },
  },
  unauthorized: {
    error: {
      code: 'INVALID_API_KEY',
      message: 'API Key inválida ou não fornecida',
    },
  },
  notFound: {
    error: {
      code: 'NOT_FOUND',
      message: 'Recurso não encontrado',
    },
  },
  tooManyRequests: {
    error: {
      code: 'RATE_LIMIT_EXCEEDED',
      message: 'Limite de requisições excedido. Aguarde 60 segundos.',
    },
  },
  internalError: {
    error: {
      code: 'INTERNAL_ERROR',
      message: 'Erro interno do servidor. Tente novamente mais tarde.',
    },
  },
};