import { ApiProperty } from '@nestjs/swagger';

export class WebhookResponseDto {
  @ApiProperty({ example: 'whk_123456' })
  id!: string;

  @ApiProperty({ example: 'https://meusistema.com/webhook' })
  url!: string;

  @ApiProperty({ example: ['payment.received', 'invoice.expired'] })
  events!: string[];

  @ApiProperty({ example: true })
  active!: boolean;

  @ApiProperty()
  created_at!: string;

  @ApiProperty()
  updated_at!: string;
}

export class WebhookListResponseDto {
  @ApiProperty({ type: [WebhookResponseDto] })
  items!: WebhookResponseDto[];

  @ApiProperty({ example: 5 })
  total!: number;
}