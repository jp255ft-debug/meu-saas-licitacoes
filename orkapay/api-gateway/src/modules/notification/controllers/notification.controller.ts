import {
  Controller,
  Post,
  Get,
  Delete,
  Param,
  Body,
  HttpCode,
  HttpStatus,
  ValidationPipe,
} from '@nestjs/common';
import {
  ApiTags,
  ApiOperation,
  ApiResponse,
  ApiParam,
  ApiHeader,
} from '@nestjs/swagger';
import { NotificationService } from '../services/notification.service';
import { CreateWebhookDto } from '../dto/create-webhook.dto';
import { WebhookResponseDto, WebhookListResponseDto } from '../dto/webhook-response.dto';
import { ErrorResponseDto, ERROR_EXAMPLES } from '../../../common/dto/error-response.dto';

@ApiTags('Webhooks (Notification)')
@ApiHeader({ name: 'X-API-Key', description: 'API Key de autenticação', required: true })
@Controller('api/v1/webhooks')
export class NotificationController {
  constructor(private readonly notificationService: NotificationService) {}

  @Post()
  @ApiOperation({ summary: 'Configurar webhook', description: 'Registra uma URL para receber notificações de eventos em tempo real.\n\nEventos disponíveis:\n- payment.received — pagamento confirmado\n- invoice.expired — fatura vencida\n- transfer.settled — transferência liquidada' })
  @ApiResponse({ status: 201, description: 'Webhook configurado com sucesso', type: WebhookResponseDto })
  @ApiResponse({ status: 400, description: 'Dados inválidos', type: ErrorResponseDto, schema: { example: { error: { code: 'INVALID_INPUT', message: 'Dados inválidos', details: [{ field: 'url', issue: 'URL inválida' }, { field: 'events', issue: 'Lista de eventos vazia' }] } } } })
  @ApiResponse({ status: 401, description: 'API Key inválida', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.unauthorized } })
  @ApiResponse({ status: 429, description: 'Limite excedido', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.tooManyRequests } })
  async createWebhook(@Body(ValidationPipe) dto: CreateWebhookDto): Promise<WebhookResponseDto> {
    return this.notificationService.createWebhook(dto);
  }

  @Get()
  @ApiOperation({ summary: 'Listar webhooks configurados', description: 'Retorna todos os webhooks configurados para a sua conta.' })
  @ApiResponse({ status: 200, description: 'Lista de webhooks', type: WebhookListResponseDto })
  @ApiResponse({ status: 401, description: 'API Key inválida', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.unauthorized } })
  async listWebhooks(): Promise<WebhookListResponseDto> {
    return this.notificationService.listWebhooks();
  }

  @Delete(':id')
  @HttpCode(HttpStatus.NO_CONTENT)
  @ApiOperation({ summary: 'Remover webhook', description: 'Remove um webhook existente. Após a remoção, não serão mais enviadas notificações para a URL.' })
  @ApiParam({ name: 'id', description: 'ID do webhook', example: 'whk_123456' })
  @ApiResponse({ status: 204, description: 'Webhook removido com sucesso' })
  @ApiResponse({ status: 404, description: 'Webhook não encontrado', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.notFound } })
  async deleteWebhook(@Param('id') id: string): Promise<void> {
    await this.notificationService.deleteWebhook(id);
  }

  @Post(':id/test')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Testar webhook', description: 'Envia um evento de teste para a URL configurada.' })
  @ApiParam({ name: 'id', description: 'ID do webhook', example: 'whk_123456' })
  @ApiResponse({ status: 200, description: 'Evento de teste enviado com sucesso', schema: { example: { status: 'success', message: 'Evento de teste enviado para https://meusistema.com/webhook' } } })
  @ApiResponse({ status: 404, description: 'Webhook não encontrado', type: ErrorResponseDto, schema: { example: ERROR_EXAMPLES.notFound } })
  @ApiResponse({ status: 400, description: 'Falha ao enviar teste', type: ErrorResponseDto, schema: { example: { error: { code: 'TEST_FAILED', message: 'Falha ao enviar evento de teste.' } } } })
  async testWebhook(@Param('id') id: string): Promise<{ status: string; message: string }> {
    return this.notificationService.testWebhook(id);
  }
}