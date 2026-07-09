import {
  Controller,
  Get,
  Param,
  Query,
  ParseIntPipe,
  DefaultValuePipe,
} from '@nestjs/common';
import {
  ApiTags,
  ApiOperation,
  ApiResponse,
  ApiParam,
  ApiQuery,
} from '@nestjs/swagger';
import { ReconciliationService } from '../services/reconciliation.service';
import {
  ReconciliationResponseDto,
  ReconciliationListResponseDto,
} from '../dto/reconciliation-response.dto';

@ApiTags('Reconciliation')
@Controller('api/v1/reconciliation')
export class ReconciliationController {
  constructor(private readonly service: ReconciliationService) {}

  @Get(':id')
  @ApiOperation({ summary: 'Buscar reconciliação por ID da invoice' })
  @ApiParam({ name: 'id', description: 'ID da invoice' })
  @ApiResponse({ status: 200, description: 'Reconciliação encontrada', type: ReconciliationResponseDto })
  @ApiResponse({ status: 404, description: 'Reconciliação não encontrada' })
  async getReconciliation(@Param('id') id: string): Promise<ReconciliationResponseDto> {
    return this.service.getReconciliation(id);
  }

  @Get()
  @ApiOperation({ summary: 'Listar reconciliações' })
  @ApiQuery({ name: 'status', required: false, description: 'Filtrar por status' })
  @ApiQuery({ name: 'limit', required: false, type: Number })
  @ApiQuery({ name: 'offset', required: false, type: Number })
  @ApiResponse({ status: 200, description: 'Lista de reconciliações', type: ReconciliationListResponseDto })
  async listReconciliations(
    @Query('status') status?: string,
    @Query('limit', new DefaultValuePipe(20), ParseIntPipe) limit?: number,
    @Query('offset', new DefaultValuePipe(0), ParseIntPipe) offset?: number,
  ): Promise<ReconciliationListResponseDto> {
    return this.service.listReconciliations(status, limit, offset);
  }
}