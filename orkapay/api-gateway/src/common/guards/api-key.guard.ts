import { Injectable, CanActivate, ExecutionContext, UnauthorizedException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Reflector } from '@nestjs/core';
import { HttpService } from '@nestjs/axios';
import { firstValueFrom } from 'rxjs';
import { IS_PUBLIC_KEY } from '../decorators/public.decorator';

@Injectable()
export class ApiKeyGuard implements CanActivate {
  constructor(
    private readonly configService: ConfigService,
    private readonly reflector: Reflector,
    private readonly httpService: HttpService,
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const isPublic = this.reflector.getAllAndOverride<boolean>(IS_PUBLIC_KEY, [
      context.getHandler(),
      context.getClass(),
    ]);
    if (isPublic) {
      return true;
    }

    const request = context.switchToHttp().getRequest();
    const apiKey = request.headers['x-api-key'] ?? request.headers['X-API-Key'];

    if (!apiKey) {
      throw new UnauthorizedException('API Key é obrigatória');
    }

    // Validar chave chamando o cashin-engine (que usa bcrypt + MapAuthProvider)
    const authServiceUrl = this.configService.get<string>('services.cashin.url') ?? 'http://localhost:8080';
    try {
      const { data, status } = await firstValueFrom(
        this.httpService.post<{ valid: boolean; client_id?: string }>(
          `${authServiceUrl}/api/v1/auth/validate`,
          { apiKey },
        ),
      );
      if (status === 200 && data?.valid) {
        // Injetar client_id no request para uso posterior
        (request as Record<string, unknown>).clientId = data.client_id;
        return true;
      }
      throw new UnauthorizedException('API Key inválida');
    } catch {
      throw new UnauthorizedException('API Key inválida');
    }
  }
}