import { Injectable, HttpException, HttpStatus, Logger } from '@nestjs/common';
import { HttpService } from '@nestjs/axios';
import { ConfigService } from '@nestjs/config';
import { firstValueFrom } from 'rxjs';
import {
  ReconciliationResponseDto,
  ReconciliationListResponseDto,
} from '../dto/reconciliation-response.dto';

@Injectable()
export class ReconciliationService {
  private readonly logger = new Logger(ReconciliationService.name);
  private readonly baseUrl: string;
  private readonly apiKey: string;
  private readonly timeout: number;

  constructor(
    private readonly httpService: HttpService,
    private readonly configService: ConfigService,
  ) {
    this.baseUrl = this.configService.get<string>('services.reconciliation.url') ?? 'http://localhost:8082';
    this.apiKey = this.configService.get<string>('services.reconciliation.apiKey') ?? '';
    this.timeout = this.configService.get<number>('services.reconciliation.timeout') ?? 30000;
  }

  private getHeaders(): Record<string, string> {
    return {
      'X-API-Key': this.apiKey,
      'Content-Type': 'application/json',
    };
  }

  private async request<T>(method: 'GET' | 'POST', path: string, data?: unknown): Promise<T> {
    try {
      const url = `${this.baseUrl}${path}`;
      const config = { headers: this.getHeaders(), timeout: this.timeout };
      const response =
        method === 'GET'
          ? await firstValueFrom(this.httpService.get<T>(url, config))
          : await firstValueFrom(this.httpService.post<T>(url, data, config));
      return response.data;
    } catch (error: unknown) {
      this.logger.error(`Erro ${method} ${path}: ${error instanceof Error ? error.message : 'Unknown'}`);
      if (error && typeof error === 'object' && 'response' in error) {
        const err = error as { response: { data?: { message?: string }; status?: number; statusText?: string } };
        const status = err.response.status ?? HttpStatus.INTERNAL_SERVER_ERROR;
        const msg = err.response.data?.message ?? err.response.statusText ?? 'Erro no serviço reconciliation';
        if (status === 404) throw new HttpException('Reconciliação não encontrada', HttpStatus.NOT_FOUND);
        if (status === 400) throw new HttpException(msg, HttpStatus.BAD_REQUEST);
        throw new HttpException(msg, status);
      }
      throw new HttpException(
        'Erro ao comunicar com o serviço de reconciliação',
        HttpStatus.INTERNAL_SERVER_ERROR,
      );
    }
  }

  async getReconciliation(id: string): Promise<ReconciliationResponseDto> {
    return this.request<ReconciliationResponseDto>('GET', `/api/v1/reconciliation/${id}`);
  }

  async listReconciliations(
    status?: string,
    limit = 20,
    offset = 0,
  ): Promise<ReconciliationListResponseDto> {
    const params = new URLSearchParams();
    if (status) params.append('status', status);
    if (limit) params.append('limit', String(Math.min(limit, 100)));
    if (offset) params.append('offset', String(offset));
    const query = params.toString() ? `?${params.toString()}` : '';
    return this.request<ReconciliationListResponseDto>('GET', `/api/v1/reconciliation${query}`);
  }
}