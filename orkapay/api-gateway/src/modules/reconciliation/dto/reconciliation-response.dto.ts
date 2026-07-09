import { ApiProperty } from '@nestjs/swagger';

export class ReconciliationResponseDto {
  @ApiProperty({ example: 'cin-20260709-abc123' })
  id!: string;

  @ApiProperty({ example: 'PAYMENT_CONFIRMED' })
  event_type!: string;

  @ApiProperty({ example: 10000 })
  amount_cents!: number;

  @ApiProperty({ example: 'PAID' })
  status!: string;

  @ApiProperty({ required: false })
  invoice_id?: string;

  @ApiProperty({ required: false })
  external_id?: string;

  @ApiProperty()
  matched_at!: string;

  @ApiProperty()
  created_at!: string;

  @ApiProperty()
  updated_at!: string;
}

export class ReconciliationListResponseDto {
  @ApiProperty({ type: [ReconciliationResponseDto] })
  items!: ReconciliationResponseDto[];

  @ApiProperty({ example: 10 })
  total!: number;

  @ApiProperty({ example: 20 })
  limit!: number;

  @ApiProperty({ example: 0 })
  offset!: number;
}