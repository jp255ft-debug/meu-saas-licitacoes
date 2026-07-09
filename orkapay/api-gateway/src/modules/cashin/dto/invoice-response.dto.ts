import { ApiProperty } from '@nestjs/swagger';

export class InvoiceResponseDto {
  @ApiProperty({ example: 'cin-20260708-abc123' })
  id!: string;

  @ApiProperty({ example: 10000 })
  amount_cents!: number;

  @ApiProperty({ enum: ['PENDING', 'ISSUED', 'PAID', 'CANCELLED', 'EXPIRED'] })
  status!: string;

  @ApiProperty({ example: '2026-08-15' })
  due_date!: string;

  @ApiProperty({ example: 'Empresa XPTO' })
  payer_name!: string;

  @ApiProperty({ example: '12.345.678/0001-99' })
  payer_tax_id!: string;

  @ApiProperty({ required: false })
  external_id?: string;

  @ApiProperty({ required: false })
  pix_qr_code?: string;

  @ApiProperty({ required: false })
  pix_key?: string;

  @ApiProperty({ required: false })
  payment_url?: string;

  @ApiProperty()
  created_at!: string;

  @ApiProperty()
  updated_at!: string;
}

export class InvoiceListResponseDto {
  @ApiProperty({ type: [InvoiceResponseDto] })
  items!: InvoiceResponseDto[];

  @ApiProperty({ example: 42 })
  total!: number;

  @ApiProperty({ example: 20 })
  limit!: number;

  @ApiProperty({ example: 0 })
  offset!: number;
}