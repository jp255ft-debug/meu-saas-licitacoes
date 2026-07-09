import { Injectable, HttpException, HttpStatus, Logger } from '@nestjs/common';
import { HttpService } from '@nestjs/axios';
import { ConfigService } from '@nestjs/config';
import { firstValueFrom } from 'rxjs';
import { CreateWebhookDto } from '../dto/create-webhook.dto';
import { WebhookResponseDto, WebhookListResponseDto } from '../dto/webhook-response.dto';

@Injectable()
export class NotificationService {
  private readonly logger = new Logger(NotificationService.name);
  private readonly baseUrl: string;
  private readonly apiKey: string;
  private readonly timeout: number;

  constructor(
    private readonly httpService: HttpService,
    private readonly configService: ConfigService,
  ) {
    this.baseUrl = this.configService.get<string>('services.notification.url') ?? 'http://localhost:8083';
    this.apiKey = this.configService.get<string>('services.notification.apiKey') ?? '';
    this.timeout = this.configService.get<number>('services.notification.timeout') ?? 30000;
  }

  private getHeaders(): Record<string, string> {
    return {
      'X-API-Key': this.apiKey,
      'Content-Type': 'application/json',
    };
  }

  private async request<T>(method: 'GET' | 'POST' | 'DELETE', path: string, data?: unknown): Promise<T> {
    try {
      const url = `${this.baseUrl}${path}`;
      const config = { headers: this.getHeaders(), timeout: this.timeout };

      let response;
      if (method === 'GET') {
        response = await firstValueFrom(this.httpService.get<T>(url, config));
      } else if (method === 'POST') {
        response = await firstValueFrom(this.httpService.post<T>(url, data, config));
      } else {
        response = await firstValueFrom(this.httpService.delete<T>(url, config));
      }

      return response.data;
    } catch (error: unknown) {
      this.logger.error(`Erro ${method} ${path}: ${error instanceof Error ? error.message : 'Unknown'}`);

      if (error && typeof error === 'object' && 'response' in error) {
        const err = error as { response: { data?: { message?: string }; status?: number; statusText?: string } };
        const status = err.response.status ?? HttpStatus.INTERNAL_SERVER_ERROR;
        const msg = err.response.data?.message ?? err.response.statusText ?? 'Erro no serviço notification';

        if (status === 404) throw new HttpException('Webhook não encontrado', HttpStatus.NOT_FOUND);
        if (status === 400) throw new HttpException(msg, HttpStatus.BAD_REQUEST);
        if (status === 401 || status === 403) throw new HttpException('Não autorizado', HttpStatus.UNAUTHORIZED);
        throw new HttpException(msg, status);
      }

      if (error && typeof error === 'object' && 'code' in error && (error as Record<string, unknown>).code === 'ECONNABORTED') {
        throw new HttpException('Tempo de espera excedido', HttpStatus.REQUEST_TIMEOUT);
      }

      throw new HttpException('Erro ao comunicar com o serviço de notificações', HttpStatus.INTERNAL_SERVER_ERROR);
    }
  }

  async createWebhook(dto: CreateWebhookDto): Promise<WebhookResponseDto> {
    return this.request<WebhookResponseDto>('POST', '/api/v1/webhooks', dto);
  }

  async listWebhooks(): Promise<WebhookListResponseDto> {
    return this.request<WebhookListResponseDto>('GET', '/api/v1/webhooks');
  }

  async deleteWebhook(id: string): Promise<void> {
    await this.request<void>('DELETE', `/api/v1/webhooks/${id}`);
  }

  async testWebhook(id: string): Promise<{ status: string; message: string }> {
    return this.request<{ status: string; message: string }>('POST', `/api/v1/webhooks/${id}/test`);
  }
}