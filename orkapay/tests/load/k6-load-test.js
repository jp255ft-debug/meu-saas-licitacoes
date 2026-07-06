// k6 Load Test for OrkaPay APIs
// Uso: k6 run tests/load/k6-load-test.js
// Para relatório HTML: k6 run --out json=results.json tests/load/k6-load-test.js

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// ---------------------------------------------------------------------------
// Métricas customizadas
// ---------------------------------------------------------------------------

const errorRate = new Rate('errors');
const invoiceCreationTrend = new Trend('invoice_creation_duration');
const transferCreationTrend = new Trend('transfer_creation_duration');
const paymentQueryTrend = new Trend('payment_query_duration');
const successCount = new Counter('successful_requests');
const failureCount = new Counter('failed_requests');

// ---------------------------------------------------------------------------
// Configuração
// ---------------------------------------------------------------------------

export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp-up inicial
    { duration: '1m', target: 20 },    // Carga moderada
    { duration: '30s', target: 50 },   // Pico
    { duration: '1m', target: 50 },    // Sustentação no pico
    { duration: '30s', target: 0 },    // Ramp-down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95% das requisições < 500ms
    http_req_failed: ['rate<0.01'],    // < 1% de falhas
    errors: ['rate<0.05'],            // < 5% de erros customizados
  },
};

// ---------------------------------------------------------------------------
// Configuração base
// ---------------------------------------------------------------------------

const BASE_URL_CASHIN = __ENV.CASHIN_URL || 'http://localhost:8080';
const BASE_URL_CASHOUT = __ENV.CASHOUT_URL || 'http://localhost:8081';
const API_KEY = __ENV.API_KEY || 'demo-key';

const params = {
  headers: {
    'X-API-Key': API_KEY,
    'Content-Type': 'application/json',
  },
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function randomCPF() {
  // Gera um CPF aleatório para testes
  const n1 = String(Math.floor(Math.random() * 9) + 1);
  const n2 = String(Math.floor(Math.random() * 9));
  const n3 = String(Math.floor(Math.random() * 9));
  const n4 = String(Math.floor(Math.random() * 9));
  const n5 = String(Math.floor(Math.random() * 9));
  const n6 = String(Math.floor(Math.random() * 9));
  const n7 = String(Math.floor(Math.random() * 9));
  const n8 = String(Math.floor(Math.random() * 9));
  const n9 = String(Math.floor(Math.random() * 9));
  return `${n1}${n2}${n3}${n4}${n5}${n6}${n7}${n8}${n9}01`;
}

function randomEmail() {
  const id = Math.random().toString(36).substring(2, 10);
  return `payer-${id}@example.com`;
}

// ---------------------------------------------------------------------------
// Cenários de teste
// ---------------------------------------------------------------------------

export default function () {
  group('CashIn - Criar Invoice', () => {
    const payload = JSON.stringify({
      amount_cents: Math.floor(Math.random() * 90000) + 1000,
      due_date: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
      payer_name: 'Cliente Teste k6',
      payer_tax_id: randomCPF(),
      payer_email: randomEmail(),
      provider: 'ASAAS',
      metadata: { source: 'k6-load-test' },
    });

    const res = http.post(`${BASE_URL_CASHIN}/api/v1/invoices`, payload, params);
    invoiceCreationTrend.add(res.timings.duration);

    const success = check(res, {
      'status é 201': (r) => r.status === 201,
      'tem external_id': (r) => JSON.parse(r.body).external_id !== undefined,
      'tem pix_qr_code': (r) => r.json('pix_qr_code') !== undefined,
    });

    if (success) {
      successCount.add(1);
    } else {
      errorRate.add(1);
      failureCount.add(1);
    }
  });

  sleep(1);

  group('CashOut - Criar Transferência', () => {
    const payload = JSON.stringify({
      amount_cents: Math.floor(Math.random() * 50000) + 500,
      payee_name: 'Beneficiário k6',
      payee_tax_id: randomCPF(),
      payee_pix_key: randomEmail(),
      payee_pix_type: 'EMAIL',
      provider: 'STARK_BANK',
      description: 'Transferência teste k6',
      metadata: { source: 'k6-load-test' },
    });

    const res = http.post(`${BASE_URL_CASHOUT}/api/v1/transfers`, payload, params);
    transferCreationTrend.add(res.timings.duration);

    const success = check(res, {
      'status é 201': (r) => r.status === 201,
      'tem external_id': (r) => JSON.parse(r.body).external_id !== undefined,
    });

    if (success) {
      successCount.add(1);
    } else {
      errorRate.add(1);
      failureCount.add(1);
    }
  });

  sleep(1);

  group('CashIn - Health Check', () => {
    const res = http.get(`${BASE_URL_CASHIN}/health`);
    paymentQueryTrend.add(res.timings.duration);

    const success = check(res, {
      'status é 200': (r) => r.status === 200,
      'body contém status': (r) => JSON.parse(r.body).status !== undefined,
    });

    if (success) {
      successCount.add(1);
    } else {
      errorRate.add(1);
      failureCount.add(1);
    }
  });

  sleep(1);

  group('CashOut - Health Check', () => {
    const res = http.get(`${BASE_URL_CASHOUT}/health`);
    paymentQueryTrend.add(res.timings.duration);

    const success = check(res, {
      'status é 200': (r) => r.status === 200,
      'body contém status': (r) => JSON.parse(r.body).status !== undefined,
    });

    if (success) {
      successCount.add(1);
    } else {
      errorRate.add(1);
      failureCount.add(1);
    }
  });
}