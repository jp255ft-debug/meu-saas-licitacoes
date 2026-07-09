export default () => ({
  port: parseInt(process.env.PORT ?? '3000', 10),
  services: {
    cashin: {
      url: process.env.CASHIN_SERVICE_URL ?? 'http://localhost:8080',
      apiKey: process.env.CASHIN_API_KEY ?? '',
      timeout: parseInt(process.env.CASHIN_TIMEOUT ?? '30000', 10),
    },
    cashout: {
      url: process.env.CASHOUT_SERVICE_URL ?? 'http://localhost:8081',
      apiKey: process.env.CASHOUT_API_KEY ?? '',
      timeout: parseInt(process.env.CASHOUT_TIMEOUT ?? '30000', 10),
    },
    payment: {
      url: process.env.PAYMENT_SERVICE_URL ?? 'http://localhost:8082',
      apiKey: process.env.PAYMENT_API_KEY ?? '',
      timeout: parseInt(process.env.PAYMENT_TIMEOUT ?? '30000', 10),
    },
    reconciliation: {
      url: process.env.RECONCILIATION_SERVICE_URL ?? 'http://localhost:8082',
      apiKey: process.env.RECONCILIATION_API_KEY ?? '',
      timeout: parseInt(process.env.RECONCILIATION_TIMEOUT ?? '30000', 10),
    },
    notification: {
      url: process.env.NOTIFICATION_SERVICE_URL ?? 'http://localhost:8083',
      apiKey: process.env.NOTIFICATION_API_KEY ?? '',
      timeout: parseInt(process.env.NOTIFICATION_TIMEOUT ?? '30000', 10),
    },
  },
  auth: {
    apiKeys: process.env.API_KEYS ? process.env.API_KEYS.split(',') : [],
  },
  throttle: {
    ttl: parseInt(process.env.THROTTLE_TTL ?? '60000', 10),
    limit: parseInt(process.env.THROTTLE_LIMIT ?? '100', 10),
  },
});
