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
} from '@nestjs/swagger';
import { NotificationService } from '../services/notification.service';
import { CreateWebhookDto } from '../dto/create-webhook.dto';
import { WebhookResponseDto, WebhookListResponseDto } from '../dto/webhook-response.dto';

@ApiTags('Webhooks (Notification)')
@Controller('api/v1/webhooks')
export class NotificationController {
  constructor(private readonly notificationService: NotificationService) {}

  @Post()
  @ApiOperation({
    summary: 'Configurar webhook',
    description: 'Registra uma URL para receber notificações de eventos (pagamento recebido, fatura vencida, etc.)',
  })
  @ApiResponse({ status: 201, description: 'Webhook configurado', type: WebhookResponseDto })
  @ApiResponse({ status: 400, description: 'Dados inválidos' })
  @ApiResponse({ status: 401, description: 'API Key inválida' })
  async createWebhook(@Body(ValidationPipe) dto: CreateWebhookDto): Promise<WebhookResponseDto> {
    return this.notificationService.createWebhook(dto);
  }

  @Get()
  @ApiOperation({ summary: 'Listar webhooks configurados' })
  @ApiResponse({ status: 200, description: 'Lista de webhooks', type: WebhookListResponseDto })
  @ApiResponse({ status: 401, description: 'API Key inválida' })
  async listWebhooks(): Promise<WebhookListResponseDto> {
    return this.notificationService.listWebhooks();
  }

  @Delete(':id')
  @HttpCode(HttpStatus.NO_CONTENT)
  @ApiOperation({ summary: 'Remover webhook' })
  @ApiParam({ name: 'id', description: 'ID do webhook' })
  @ApiResponse({ status: 204, description: 'Webhook removido' })
  @ApiResponse({ status: 404, description: 'Webhook não encontrado' })
  async deleteWebhook(@Param('id') id: string): Promise<void> {
    await this.notificationService.deleteWebhook(id);
  }

  @Post(':id/test')
  @HttpCode(HttpStatus.OK)
  @ApiOperation({ summary: 'Testar webhook', description: 'Envia um evento de teste para a URL configurada' })
  @ApiParam({ name: 'id', description: 'ID do webhook' })
  @ApiResponse({ status: 200, description: 'Evento de teste enviado' })
  @ApiResponse({ status: 404, description: 'Webhook não encontrado' })
  async testWebhook(@Param('id') id: string): Promise<{ status: string; message: string }> {
    return this.notificationService.testWebhook(id);
  }
}