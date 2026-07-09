import { Controller, Post, Get, Param, Body, Query, HttpCode, HttpStatus, ValidationPipe, ParseIntPipe, DefaultValuePipe } from '@nestjs/common';
import { ApiTags, ApiOperation, ApiResponse, ApiParam, ApiQuery, ApiHeader } from '@nestjs/swagger';
import { CashoutService } from '../services/cashout.service';
import { CreateTransferDto } from '../dto/create-transfer.dto';
import { TransferResponseDto, TransferListResponseDto } from '../dto/transfer-response.dto';
import { ErrorResponseDto, ERROR_EXAMPLES } from '../../../common/dto/error-response.dto';

@ApiTags('Cash-out (Transfers)')
@ApiHeader({ name: 'X-API-Key', description: 'API Key de autenticação', required: true })
@Controller('api/v1/transfers')
export class CashoutController {
  constructor(private readonly cashoutService: CashoutService) {}

  @Post()
  @ApiOperation({ summary: 'Criar transferência', description: 'Solicita uma transferência Pix. A transferência inicia como PROCESSING e precisa ser aprovada.' })
  @ApiResponse({ status: 201, description: 'Transferência criada', type: TransferResponseDto })
  @ApiResponse({ status: 400, description: 'Dados inválidos', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.badRequest } })
  @ApiResponse({ status: 401, description: 'Não autorizado', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.unauthorized } })
  @ApiResponse({ status: 429, description: 'Limite excedido', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.tooManyRequests } })
  async createTransfer(@Body(ValidationPipe) dto: CreateTransferDto): Promise<TransferResponseDto> {
    return this.cashoutService.createTransfer(dto);
  }

  @Get(':id')
  @ApiOperation({ summary: 'Buscar transferência por ID', description: 'Retorna detalhes de uma transferência.' })
  @ApiParam({ name: 'id', description: 'ID da transferência' })
  @ApiResponse({ status: 200, description: 'Transferência encontrada', type: TransferResponseDto })
  @ApiResponse({ status: 404, description: 'Transferência não encontrada', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.notFound } })
  @ApiResponse({ status: 401, description: 'Não autorizado', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.unauthorized } })
  async getTransfer(@Param('id') id: string): Promise<TransferResponseDto> {
    return this.cashoutService.getTransfer(id);
  }

  @Get()
  @ApiOperation({ summary: 'Listar transferências', description: 'Lista transferências com filtros opcionais.' })
  @ApiQuery({ name: 'status', required: false, enum: ['PROCESSING', 'APPROVED', 'SETTLED', 'FAILED', 'CANCELLED'] })
  @ApiQuery({ name: 'limit', required: false, type: Number, description: 'Itens por página (máx 100)' })
  @ApiQuery({ name: 'offset', required: false, type: Number })
  @ApiResponse({ status: 200, description: 'Lista de transferências', type: TransferListResponseDto })
  @ApiResponse({ status: 401, description: 'Não autorizado', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.unauthorized } })
  async listTransfers(
    @Query('status') status?: string,
    @Query('limit', new DefaultValuePipe(20), ParseIntPipe) limit?: number,
    @Query('offset', new DefaultValuePipe(0), ParseIntPipe) offset?: number,
  ): Promise<TransferListResponseDto> {
    return this.cashoutService.listTransfers(status, limit, offset);
  }

  @Post(':id/approve')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Aprovar transferência', description: 'Aprova uma transferência para liquidação.' })
  @ApiParam({ name: 'id', description: 'ID da transferência' })
  @ApiResponse({ status: 200, description: 'Transferência aprovada', type: TransferResponseDto })
  @ApiResponse({ status: 404, description: 'Transferência não encontrada', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.notFound } })
  async approveTransfer(@Param('id') id: string): Promise<TransferResponseDto> {
    return this.cashoutService.approveTransfer(id);
  }

  @Post(':id/settle')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Liquidar transferência', description: 'Executa o Pix via provedor (Stark Bank).' })
  @ApiParam({ name: 'id', description: 'ID da transferência' })
  @ApiResponse({ status: 200, description: 'Transferência liquidada', type: TransferResponseDto })
  @ApiResponse({ status: 404, description: 'Transferência não encontrada', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.notFound } })
  async settleTransfer(@Param('id') id: string): Promise<TransferResponseDto> {
    return this.cashoutService.settleTransfer(id);
  }

  @Post(':id/fail')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Falhar transferência', description: 'Marca a transferência como falha.' })
  @ApiParam({ name: 'id', description: 'ID da transferência' })
  @ApiResponse({ status: 200, description: 'Transferência falhou', type: TransferResponseDto })
  @ApiResponse({ status: 404, description: 'Transferência não encontrada', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.notFound } })
  async failTransfer(@Param('id') id: string): Promise<TransferResponseDto> {
    return this.cashoutService.failTransfer(id);
  }

  @Post(':id/cancel')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Cancelar transferência', description: 'Cancela uma transferência antes da liquidação.' })
  @ApiParam({ name: 'id', description: 'ID da transferência' })
  @ApiResponse({ status: 200, description: 'Transferência cancelada', type: TransferResponseDto })
  @ApiResponse({ status: 404, description: 'Transferência não encontrada', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.notFound } })
  async cancelTransfer(@Param('id') id: string): Promise<TransferResponseDto> {
    return this.cashoutService.cancelTransfer(id);
  }
}