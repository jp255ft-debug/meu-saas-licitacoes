import { CallHandler, ExecutionContext, Injectable, NestInterceptor, HttpException, HttpStatus } from '@nestjs/common';
import { Observable, of } from 'rxjs';
import { tap } from 'rxjs/operators';

@Injectable()
export class IdempotencyInterceptor implements NestInterceptor {
  private readonly ttlSeconds = 86400; // 24h

  async intercept(context: ExecutionContext, next: CallHandler): Promise<Observable<any>> {
    const request = context.switchToHttp().getRequest();
    const idempotencyKey = request.headers['x-idempotency-key'];

    // Requer chave de idempotência para métodos mutantes
    if (['POST', 'PUT', 'PATCH'].includes(request.method) && !idempotencyKey) {
      throw new HttpException(
        'Idempotency-Key header is required for state-changing operations',
        HttpStatus.BAD_REQUEST,
      );
    }

    if (!idempotencyKey) {
      return next.handle();
    }

    // TODO: Implementar com Redis real
    // const cachedResponse = await this.redisService.get(idempotencyKey);
    const cachedResponse = null;

    if (cachedResponse) {
      return of(JSON.parse(cachedResponse));
    }

    return next.handle().pipe(
      tap((response) => {
        // TODO: Salvar no Redis com TTL de 24h
        // this.redisService.set(idempotencyKey, JSON.stringify(response), 'EX', this.ttlSeconds);
        console.log(`[Idempotency] Key: ${idempotencyKey} - Response cached for ${this.ttlSeconds}s`);
      }),
    );
  }
}