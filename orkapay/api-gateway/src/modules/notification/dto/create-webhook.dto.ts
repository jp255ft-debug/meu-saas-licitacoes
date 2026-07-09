import { ApiProperty } from '@nestjs/swagger';
import { IsString, IsUrl, IsArray, ArrayNotEmpty, IsOptional } from 'class-validator';

export class CreateWebhookDto {
  @ApiProperty({ example: 'https://meusistema.com/webhook', description: 'URL que receberá as notificações' })
  @IsUrl()
  url!: string;

  @ApiProperty({
    example: ['payment.received', 'invoice.expired'],
    description: 'Lista de eventos que deseja receber',
    enum: ['payment.received', 'invoice.expired', 'transfer.settled'],
  })
  @IsArray()
  @ArrayNotEmpty()
  events!: string[];

  @ApiProperty({
    required: false,
    example: 'meu-token-secreto',
    description: 'Token para autenticação do webhook',
  })
  @IsOptional()
  @IsString()
  secret?: string;
}