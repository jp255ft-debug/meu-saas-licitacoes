import { Controller, Post, Get, Param, Body, Query, HttpCode, HttpStatus, ValidationPipe, ParseIntPipe, DefaultValuePipe } from '@nestjs/common';
import { ApiTags, ApiOperation, ApiResponse, ApiParam, ApiQuery, ApiHeader } from '@nestjs/swagger';
import { CashinService } from '../services/cashin.service';
import { CreateInvoiceDto } from '../dto/create-invoice.dto';
import { InvoiceResponseDto, InvoiceListResponseDto } from '../dto/invoice-response.dto';
import { ErrorResponseDto, ERROR_EXAMPLES } from '../../../common/dto/error-response.dto';

@ApiTags('Cash-in (Invoices)')
@ApiHeader({ name: 'X-API-Key', description: 'API Key de autenticação', required: true })
@Controller('api/v1/invoices')
export class CashinController {
  constructor(private readonly cashinService: CashinService) {}

  @Post()
  @ApiOperation({
    summary: 'Criar uma nova fatura',
    description: 'Cria uma fatura para cobrança B2B. A fatura inicia com status PENDING.\n\n**Fluxo completo:**\n1. Crie a fatura (PENDING)\n2. Emita com POST /invoices/{id}/issue (gera PIX/boleto)\n3. Cliente paga\n4. Webhook atualiza para PAID',
  })
  @ApiResponse({ status: 201, description: 'Fatura criada com sucesso', type: InvoiceResponseDto, schema: { example: { id: 'cin-20260709-abc123', amount_cents: 10000, status: 'PENDING', due_date: '2026-08-15', payer_name: 'Empresa XPTO', payer_tax_id: '11222333000181', external_id: null, pix_qr_code: null, pix_key: null, provider: null, created_at: '2026-07-09T10:00:00Z', updated_at: '2026-07-09T10:00:00Z' } } })
  @ApiResponse({ status: 400, description: 'Dados inválidos', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.badRequest } })
  @ApiResponse({ status: 401, description: 'API Key inválida', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.unauthorized } })
  @ApiResponse({ status: 429, description: 'Limite de requisições excedido', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.tooManyRequests } })
  async createInvoice(@Body(ValidationPipe) dto: CreateInvoiceDto): Promise<InvoiceResponseDto> {
    return this.cashinService.createInvoice(dto);
  }

  @Get(':id')
  @ApiOperation({ summary: 'Buscar fatura por ID', description: 'Retorna os detalhes de uma fatura específica.' })
  @ApiParam({ name: 'id', description: 'ID da fatura (formato: cin-YYYYMMDD-xxxx)', example: 'cin-20260709-abc123' })
  @ApiResponse({ status: 200, description: 'Fatura encontrada', type: InvoiceResponseDto, schema: { example: { id: 'cin-20260709-abc123', amount_cents: 10000, status: 'PAID', due_date: '2026-08-15', payer_name: 'Empresa XPTO', payer_tax_id: '11222333000181', external_id: 'pay_abc123xyz', pix_qr_code: '000201010212...', pix_key: 'pixkey@orkapay.com', provider: 'ASAAS', created_at: '2026-07-09T10:00:00Z', updated_at: '2026-07-09T14:30:00Z' } } })
  @ApiResponse({ status: 404, description: 'Fatura não encontrada', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.notFound } })
  @ApiResponse({ status: 401, description: 'API Key inválida', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.unauthorized } })
  async getInvoice(@Param('id') id: string): Promise<InvoiceResponseDto> {
    return this.cashinService.getInvoice(id);
  }

  @Get()
  @ApiOperation({ summary: 'Listar faturas', description: 'Lista faturas com filtros opcionais por status, paginação.' })
  @ApiQuery({ name: 'status', required: false, description: 'Filtrar por status', enum: ['PENDING', 'ISSUED', 'PAID', 'CANCELLED', 'EXPIRED'] })
  @ApiQuery({ name: 'limit', required: false, type: Number, description: 'Itens por página (máx 100)' })
  @ApiQuery({ name: 'offset', required: false, type: Number, description: 'Offset para paginação' })
  @ApiResponse({ status: 200, description: 'Lista de faturas', type: InvoiceListResponseDto })
  @ApiResponse({ status: 401, description: 'API Key inválida', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.unauthorized } })
  async listInvoices(
    @Query('status') status?: string,
    @Query('limit', new DefaultValuePipe(20), ParseIntPipe) limit?: number,
    @Query('offset', new DefaultValuePipe(0), ParseIntPipe) offset?: number,
  ): Promise<InvoiceListResponseDto> {
    return this.cashinService.listInvoices(status, limit, offset);
  }

  @Post(':id/issue')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Emitir fatura (gerar PIX/boleto)', description: 'Emite uma fatura pendente, gerando QR Code PIX, chave PIX e/ou boleto.' })
  @ApiParam({ name: 'id', description: 'ID da fatura' })
  @ApiResponse({ status: 200, description: 'Fatura emitida com PIX/boleto', type: InvoiceResponseDto })
  @ApiResponse({ status: 404, description: 'Fatura não encontrada', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.notFound } })
  @ApiResponse({ status: 400, description: 'Fatura não pode ser emitida', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.badRequest } })
  async issueInvoice(@Param('id') id: string): Promise<InvoiceResponseDto> {
    return this.cashinService.issueInvoice(id);
  }

  @Post(':id/pay')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Registrar pagamento manual', description: 'Marca uma fatura como paga (PAID). Útil para correção manual ou simulação.' })
  @ApiParam({ name: 'id', description: 'ID da fatura' })
  @ApiResponse({ status: 200, description: 'Fatura paga', type: InvoiceResponseDto })
  @ApiResponse({ status: 404, description: 'Fatura não encontrada', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.notFound } })
  @ApiResponse({ status: 400, description: 'Fatura não pode ser paga', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.badRequest } })
  async payInvoice(@Param('id') id: string): Promise<InvoiceResponseDto> {
    return this.cashinService.payInvoice(id);
  }

  @Post(':id/cancel')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Cancelar fatura', description: 'Cancela uma fatura pendente ou emitida.' })
  @ApiParam({ name: 'id', description: 'ID da fatura' })
  @ApiResponse({ status: 200, description: 'Fatura cancelada', type: InvoiceResponseDto })
  @ApiResponse({ status: 404, description: 'Fatura não encontrada', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.notFound } })
  @ApiResponse({ status: 400, description: 'Fatura não pode ser cancelada', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.badRequest } })
  async cancelInvoice(@Param('id') id: string): Promise<InvoiceResponseDto> {
    return this.cashinService.cancelInvoice(id);
  }

  @Post(':id/expire')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Expirar fatura', description: 'Marca uma fatura pendente como expirada.' })
  @ApiParam({ name: 'id', description: 'ID da fatura' })
  @ApiResponse({ status: 200, description: 'Fatura expirada', type: InvoiceResponseDto })
  @ApiResponse({ status: 404, description: 'Fatura não encontrada', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.notFound } })
  @ApiResponse({ status: 400, description: 'Fatura não pode ser expirada', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.badRequest } })
  async expireInvoice(@Param('id') id: string): Promise<InvoiceResponseDto> {
    return this.cashinService.expireInvoice(id);
  }
}