import {
  Controller,
  Get,
  Param,
  Query,
  DefaultValuePipe,
  ParseIntPipe,
} from '@nestjs/common';
import {
  ApiTags,
  ApiOperation,
  ApiResponse,
  ApiParam,
  ApiQuery,
  ApiHeader,
} from '@nestjs/swagger';
import { ReconciliationService } from '../services/reconciliation.service';
import {
  ReconciliationResponseDto,
  ReconciliationListResponseDto,
} from '../dto/reconciliation-response.dto';
import { ErrorResponseDto, ERROR_EXAMPLES } from '../../../common/dto/error-response.dto';

@ApiTags('Reconciliation')
@ApiHeader({ name: 'X-API-Key', description: 'API Key de autenticação', required: true })
@Controller('api/v1/reconciliation')
export class ReconciliationController {
  constructor(private readonly reconciliationService: ReconciliationService) {}

  @Get(':id')
  @ApiOperation({ summary: 'Buscar reconciliação por ID da invoice', description: 'Retorna os detalhes de uma reconciliação específica.' })
  @ApiParam({ name: 'id', description: 'ID da fatura', example: 'cin-20260709-abc123' })
  @ApiResponse({ status: 200, description: 'Reconciliação encontrada', type: ReconciliationResponseDto })
  @ApiResponse({ status: 404, description: 'Reconciliação não encontrada', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.notFound } })
  @ApiResponse({ status: 401, description: 'Não autorizado', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.unauthorized } })
  async getReconciliation(@Param('id') id: string): Promise<ReconciliationResponseDto> {
    return this.reconciliationService.getReconciliation(id);
  }

  @Get()
  @ApiOperation({ summary: 'Listar reconciliações', description: 'Retorna lista paginada de reconciliações com filtros opcionais.' })
  @ApiQuery({ name: 'status', required: false, description: 'Filtrar por status' })
  @ApiQuery({ name: 'limit', required: false, type: Number, description: 'Itens por página (máx 100)' })
  @ApiQuery({ name: 'offset', required: false, type: Number, description: 'Offset para paginação' })
  @ApiResponse({ status: 200, description: 'Lista de reconciliações', type: ReconciliationListResponseDto })
  @ApiResponse({ status: 401, description: 'Não autorizado', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.unauthorized } })
  async listReconciliations(
    @Query('status') status?: string,
    @Query('limit', new DefaultValuePipe(20), ParseIntPipe) limit?: number,
    @Query('offset', new DefaultValuePipe(0), ParseIntPipe) offset?: number,
  ): Promise<ReconciliationListResponseDto> {
    return this.reconciliationService.listReconciliations(status, limit, offset);
  }
}