// k6 Security Test for OrkaPay
// Testa payloads maliciosos: SQL Injection, XSS, overflow, campos vazios
// Uso: k6 run tests/load/k6-security-test.js

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate } from 'k6/metrics';

// ---------------------------------------------------------------------------
// Métricas
// ---------------------------------------------------------------------------

const errorRate = new Rate('errors');
const maliciousBlockRate = new Rate('malicious_blocked');

// ---------------------------------------------------------------------------
// Configuração
// ---------------------------------------------------------------------------

export const options = {
  stages: [
    { duration: '10s', target: 5 },
    { duration: '20s', target: 5 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.10'],  // Até 10% de falhas aceitável (payloads inválidos)
    malicious_blocked: ['rate>0.80'], // Pelo menos 80% dos payloads maliciosos devem ser rejeitados
  },
};

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
// Payloads maliciosos para CashIn
// ---------------------------------------------------------------------------

const maliciousInvoices = [
  // SQL Injection
  {
    name: 'SQL Injection no tax_id',
    payload: {
      amount_cents: 10000,
      due_date: '2026-07-10',
      payer_name: 'Hacker',
      payer_tax_id: "' OR '1'='1",
      payer_email: "' OR '1'='1@test.com",
      provider: 'ASAAS',
    },
    expected_status: 400,
  },
  // XSS
  {
    name: 'XSS no payer_name',
    payload: {
      amount_cents: 10000,
      due_date: '2026-07-10',
      payer_name: '<script>alert(1)</script>',
      payer_tax_id: '12345678901',
      payer_email: '<script>alert(1)</script>',
      provider: 'ASAAS',
    },
    expected_status: 400,
  },
  // Overflow de amount
  {
    name: 'Overflow de amount_cents',
    payload: {
      amount_cents: 999999999999999999,
      due_date: '2026-07-10',
      payer_name: 'João',
      payer_tax_id: '12345678901',
      provider: 'ASAAS',
    },
    expected_status: 400,
  },
  // Amount negativo
  {
    name: 'Amount negativo',
    payload: {
      amount_cents: -1000,
      due_date: '2026-07-10',
      payer_name: 'João',
      payer_tax_id: '12345678901',
      provider: 'ASAAS',
    },
    expected_status: 400,
  },
  // Campos vazios
  {
    name: 'Campos obrigatórios vazios',
    payload: {
      amount_cents: 0,
      due_date: '',
      payer_name: '',
      payer_tax_id: '',
      provider: '',
    },
    expected_status: 400,
  },
  // JSON malformado (payload como string)
  {
    name: 'Payload não-JSON',
    payload: 'not-a-json-string',
    expected_status: 400,
  },
  // Injeção de comando
  {
    name: 'Command injection',
    payload: {
      amount_cents: 10000,
      due_date: '2026-07-10',
      payer_name: 'João; rm -rf /',
      payer_tax_id: '12345678901; drop table invoices',
      provider: 'ASAAS',
    },
    expected_status: 400,
  },
  // Payload gigante
  {
    name: 'Payload extremamente grande',
    payload: {
      amount_cents: 10000,
      due_date: '2026-07-10',
      payer_name: 'A'.repeat(10000),
      payer_tax_id: '1'.repeat(10000),
      provider: 'ASAAS',
    },
    expected_status: 400,
  },
];

// ---------------------------------------------------------------------------
// Payloads maliciosos para CashOut
// ---------------------------------------------------------------------------

const maliciousTransfers = [
  {
    name: 'SQL Injection CashOut',
    payload: {
      amount_cents: 50000,
      payee_name: "'; DROP TABLE transfers; --",
      payee_tax_id: "' OR '1'='1",
      payee_pix_key: "' OR '1'='1@test.com",
      payee_pix_type: "EMAIL'; DROP TABLE users; --",
      provider: 'STARK_BANK',
      description: 'Test',
    },
    expected_status: 400,
  },
  {
    name: 'Overflow CashOut',
    payload: {
      amount_cents: 999999999999999999,
      payee_name: 'Jane',
      payee_tax_id: '98765432101',
      payee_pix_key: 'jane@test.com',
      payee_pix_type: 'EMAIL',
      provider: 'STARK_BANK',
      description: 'Test',
    },
    expected_status: 400,
  },
  {
    name: 'XSS CashOut',
    payload: {
      amount_cents: 50000,
      payee_name: '<img src=x onerror=alert(1)>',
      payee_tax_id: '98765432101',
      payee_pix_key: 'jane@test.com',
      payee_pix_type: 'EMAIL',
      provider: '<script>alert(1)</script>',
      description: '<script>alert(1)</script>',
    },
    expected_status: 400,
  },
];

// ---------------------------------------------------------------------------
// Cenários de teste
// ---------------------------------------------------------------------------

export default function () {
  let totalRequests = 0;
  let blockedRequests = 0;

  group('CashIn - Payloads Maliciosos', () => {
    maliciousInvoices.forEach((testCase) => {
      const payload = typeof testCase.payload === 'string'
        ? testCase.payload
        : JSON.stringify(testCase.payload);

      const res = http.post(`${BASE_URL_CASHIN}/api/v1/invoices`, payload, params);
      totalRequests++;

      // Verifica se foi rejeitado (400 ou 422)
      const isBlocked = res.status === 400 || res.status === 422 || res.status === 413;
      if (isBlocked) {
        blockedRequests++;
      }

      check(res, {
        [`${testCase.name} - rejeitado (esperado ${testCase.expected_status})`]:
          (r) => r.status === testCase.expected_status || isBlocked,
      });

      sleep(0.5);
    });
  });

  group('CashOut - Payloads Maliciosos', () => {
    maliciousTransfers.forEach((testCase) => {
      const payload = JSON.stringify(testCase.payload);
      const res = http.post(`${BASE_URL_CASHOUT}/api/v1/transfers`, payload, params);
      totalRequests++;

      const isBlocked = res.status === 400 || res.status === 422;
      if (isBlocked) {
        blockedRequests++;
      }

      check(res, {
        [`${testCase.name} - rejeitado (esperado ${testCase.expected_status})`]:
          (r) => r.status === testCase.expected_status || isBlocked,
      });

      sleep(0.5);
    });
  });

  group('Autenticação - API Key Inválida', () => {
    const payload = JSON.stringify({
      amount_cents: 10000,
      due_date: '2026-07-10',
      payer_name: 'João',
      payer_tax_id: '12345678901',
      provider: 'ASAAS',
    });

    const invalidParams = {
      headers: {
        'X-API-Key': 'invalid-key-12345',
        'Content-Type': 'application/json',
      },
    };

    const res = http.post(`${BASE_URL_CASHIN}/api/v1/invoices`, payload, invalidParams);
    totalRequests++;

    check(res, {
      'API Key inválida → 401': (r) => r.status === 401,
    });
  });

  // Atualiza métrica
  if (totalRequests > 0) {
    maliciousBlockRate.add(blockedRequests / totalRequests);
  }
}