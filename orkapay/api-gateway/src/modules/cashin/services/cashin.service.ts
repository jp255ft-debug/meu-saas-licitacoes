import { Injectable, HttpException, HttpStatus, Logger } from '@nestjs/common';
import { HttpService } from '@nestjs/axios';
import { ConfigService } from '@nestjs/config';
import { firstValueFrom } from 'rxjs';
import { CreateInvoiceDto } from '../dto/create-invoice.dto';
import { InvoiceResponseDto, InvoiceListResponseDto } from '../dto/invoice-response.dto';

@Injectable()
export class CashinService {
  private readonly logger = new Logger(CashinService.name);
  private readonly baseUrl: string;
  private readonly apiKey: string;
  private readonly timeout: number;

  constructor(
    private readonly httpService: HttpService,
    private readonly configService: ConfigService,
  ) {
    this.baseUrl = this.configService.get<string>('services.cashin.url') ?? 'http://localhost:8080';
    this.apiKey = this.configService.get<string>('services.cashin.apiKey') ?? '';
    this.timeout = this.configService.get<number>('services.cashin.timeout') ?? 30000;
  }

  private getHeaders(): Record<string, string> {
    return {
      'X-API-Key': this.apiKey,
      'Content-Type': 'application/json',
    };
  }

  private async request<T>(method: string, path: string, data?: unknown): Promise<T> {
    try {
      const url = `${this.baseUrl}${path}`;
      const config = { 
        headers: this.getHeaders(),
        timeout: this.timeout,
      };
      let response;

      switch (method) {
        case 'GET':
          response = await firstValueFrom(this.httpService.get<T>(url, config));
          break;
        case 'POST':
          response = await firstValueFrom(this.httpService.post<T>(url, data, config));
          break;
        default:
          throw new Error(`Método ${method} não suportado`);
      }

      return response.data;
    } catch (error: unknown) {
      if (error && typeof error === 'object') {
        const err = error as Record<string, unknown>;
        
        // Timeout do axios
        if (err.code === 'ECONNABORTED') {
          this.logger.error(`Timeout na requisição ${method} ${path}`);
          throw new HttpException(
            'Tempo de espera excedido ao comunicar com o serviço de faturas',
            HttpStatus.REQUEST_TIMEOUT,
          );
        }

        if ('response' in err) {
          const axiosErr = err as { 
            response?: { 
              data?: Record<string, unknown>; 
              status?: number; 
              statusText?: string 
            } 
          };
          const status = axiosErr.response?.status ?? HttpStatus.INTERNAL_SERVER_ERROR;
          const msg = 
            (axiosErr.response?.data?.message as string) ?? 
            axiosErr.response?.statusText ?? 
            'Erro no serviço cashin';

          if (status === 404) throw new HttpException('Fatura não encontrada', HttpStatus.NOT_FOUND);
          if (status === 400) throw new HttpException(msg, HttpStatus.BAD_REQUEST);
          if (status === 401 || status === 403) throw new HttpException('Não autorizado', HttpStatus.UNAUTHORIZED);
          if (status === 429) throw new HttpException('Limite de requisições excedido', HttpStatus.TOO_MANY_REQUESTS);
          throw new HttpException(msg, status);
        }
      }

      this.logger.error(`Erro ao comunicar com cashin: ${JSON.stringify(error)}`);
      throw new HttpException(
        'Erro ao comunicar com o serviço de faturas',
        HttpStatus.INTERNAL_SERVER_ERROR,
      );
    }
  }

  async createInvoice(dto: CreateInvoiceDto): Promise<InvoiceResponseDto> {
    return this.request<InvoiceResponseDto>('POST', '/api/v1/invoices', dto);
  }

  async getInvoice(id: string): Promise<InvoiceResponseDto> {
    return this.request<InvoiceResponseDto>('GET', `/api/v1/invoices/${id}`);
  }

  async listInvoices(
    status?: string,
    limit?: number,
    offset?: number,
  ): Promise<InvoiceListResponseDto> {
    const params = new URLSearchParams();
    if (status) params.append('status', status);
    if (limit) params.append('limit', String(Math.min(limit ?? 20, 100)));
    if (offset) params.append('offset', String(offset ?? 0));
    const query = params.toString() ? `?${params.toString()}` : '';
    return this.request<InvoiceListResponseDto>('GET', `/api/v1/invoices${query}`);
  }

  async issueInvoice(id: string): Promise<InvoiceResponseDto> {
    return this.request<InvoiceResponseDto>('POST', `/api/v1/invoices/${id}/issue`);
  }

  async payInvoice(id: string): Promise<InvoiceResponseDto> {
    return this.request<InvoiceResponseDto>('POST', `/api/v1/invoices/${id}/pay`);
  }

  async cancelInvoice(id: string): Promise<InvoiceResponseDto> {
    return this.request<InvoiceResponseDto>('POST', `/api/v1/invoices/${id}/cancel`);
  }

  async expireInvoice(id: string): Promise<InvoiceResponseDto> {
    return this.request<InvoiceResponseDto>('POST', `/api/v1/invoices/${id}/expire`);
  }
}