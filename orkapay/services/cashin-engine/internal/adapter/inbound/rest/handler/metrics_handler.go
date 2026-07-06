package handler

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
)

// Metrics holds all application metrics as atomic counters (Prometheus format).
type Metrics struct {
	mu                      sync.Mutex
	requestsTotal           map[string]*int64
	invoicesCreated         int64
	invoicesPaid            int64
	invoicesExpired         int64
	invoicesCanceled        int64
	webhookDeliveries       int64
	webhookDeliveriesFailed int64
}

// NewMetrics creates a new metrics collector.
func NewMetrics() *Metrics {
	return &Metrics{
		requestsTotal: make(map[string]*int64),
	}
}

func (m *Metrics) getOrCreateCounter(counterMap map[string]*int64, key string) *int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := counterMap[key]; ok {
		return c
	}
	var v int64
	counterMap[key] = &v
	return &v
}

// IncRequest increments request counter for method+path+status.
func (m *Metrics) IncRequest(method, path, status string) {
	key := method + "|" + path + "|" + status
	atomic.AddInt64(m.getOrCreateCounter(m.requestsTotal, key), 1)
}

func (m *Metrics) IncInvoicesCreated()       { atomic.AddInt64(&m.invoicesCreated, 1) }
func (m *Metrics) IncInvoicesPaid()          { atomic.AddInt64(&m.invoicesPaid, 1) }
func (m *Metrics) IncInvoicesExpired()       { atomic.AddInt64(&m.invoicesExpired, 1) }
func (m *Metrics) IncInvoicesCanceled()      { atomic.AddInt64(&m.invoicesCanceled, 1) }
func (m *Metrics) IncWebhookDeliveries()     { atomic.AddInt64(&m.webhookDeliveries, 1) }
func (m *Metrics) IncWebhookDeliveriesFailed() { atomic.AddInt64(&m.webhookDeliveriesFailed, 1) }

// MetricsHandler serves metrics in Prometheus text format.
func (m *Metrics) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintln(w, "# HELP orkapay_requests_total Total HTTP requests")
	fmt.Fprintln(w, "# TYPE orkapay_requests_total counter")
	m.mu.Lock()
	for key, counter := range m.requestsTotal {
		parts := splitKey(key)
		val := atomic.LoadInt64(counter)
		fmt.Fprintf(w, `orkapay_requests_total{method="%s",path="%s",status="%s"} %d`+"\n", parts[0], parts[1], parts[2], val)
	}
	m.mu.Unlock()
	fmt.Fprintln(w, "\n# HELP orkapay_invoices_created_total Total invoices created")
	fmt.Fprintln(w, "# TYPE orkapay_invoices_created_total counter")
	fmt.Fprintf(w, "orkapay_invoices_created_total %d\n", atomic.LoadInt64(&m.invoicesCreated))
	fmt.Fprintln(w, "\n# HELP orkapay_invoices_paid_total Total invoices paid")
	fmt.Fprintln(w, "# TYPE orkapay_invoices_paid_total counter")
	fmt.Fprintf(w, "orkapay_invoices_paid_total %d\n", atomic.LoadInt64(&m.invoicesPaid))
	fmt.Fprintln(w, "\n# HELP orkapay_invoices_expired_total Total invoices expired")
	fmt.Fprintln(w, "# TYPE orkapay_invoices_expired_total counter")
	fmt.Fprintf(w, "orkapay_invoices_expired_total %d\n", atomic.LoadInt64(&m.invoicesExpired))
	fmt.Fprintln(w, "\n# HELP orkapay_webhook_deliveries_total Total webhook deliveries")
	fmt.Fprintln(w, "# TYPE orkapay_webhook_deliveries_total counter")
	fmt.Fprintf(w, "orkapay_webhook_deliveries_total %d\n", atomic.LoadInt64(&m.webhookDeliveries))
}

func splitKey(key string) []string {
	parts := make([]string, 3)
	current := 0
	start := 0
	for i := 0; i < len(key) && current < 2; i++ {
		if key[i] == '|' {
			parts[current] = key[start:i]
			current++
			start = i + 1
		}
	}
	parts[current] = key[start:]
	return parts
}