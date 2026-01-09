package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/url"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
)

var pollCount int64
var runtimeMetricsGauge atomic.Value
var runtimeMetricsCounter atomic.Value

var (
	pollInterval   = flag.Duration("poll-interval", 2*time.Second, "poll interval")
	reportInterval = flag.Duration("report-interval", 10*time.Second, "report interval")
)

type RuntimeMetrics map[string]float64
type CustomMetrics map[string]int64

type MetricsClient struct {
	baseURL string
	client  *resty.Client
}

func NewMetricsClient(baseURL string) *MetricsClient {
	client := resty.New().
		SetTimeout(3*time.Second).
		SetHeader("Content-Type", "text/plain")
	return &MetricsClient{
		baseURL: baseURL,
		client:  client,
	}
}

func CollectRuntimeMetrics() RuntimeMetrics {
	var ms runtime.MemStats

	// Снимаем снимок состояния памяти рантайма
	runtime.ReadMemStats(&ms)

	// Заполняем map метрик
	return RuntimeMetrics{
		// ===== Основные аллокации =====

		"Alloc":      float64(ms.Alloc),      // текущие байты, выделенные под heap-объекты
		"TotalAlloc": float64(ms.TotalAlloc), // всего выделено байт за всё время
		"Sys":        float64(ms.Sys),        // всего байт, полученных от ОС
		"Lookups":    float64(ms.Lookups),    // количество pointer lookups
		"Mallocs":    float64(ms.Mallocs),    // количество malloc
		"Frees":      float64(ms.Frees),      // количество free

		// ===== Heap =====

		"HeapAlloc":    float64(ms.HeapAlloc),    // байты, выделенные под heap
		"HeapSys":      float64(ms.HeapSys),      // байты, запрошенные у ОС под heap
		"HeapIdle":     float64(ms.HeapIdle),     // неиспользуемые байты heap
		"HeapInuse":    float64(ms.HeapInuse),    // используемые байты heap
		"HeapReleased": float64(ms.HeapReleased), // байты heap, возвращённые ОС
		"HeapObjects":  float64(ms.HeapObjects),  // количество объектов в heap

		// ===== GC =====

		"GCCPUFraction": float64(ms.GCCPUFraction), // доля CPU, потраченная на GC
		"GCSys":         float64(ms.GCSys),         // байты под GC структуры
		"NextGC":        float64(ms.NextGC),        // порог heap для следующего GC
		"LastGC":        float64(ms.LastGC),        // timestamp последнего GC (ns)
		"PauseTotalNs":  float64(ms.PauseTotalNs),  // суммарные паузы GC
		"NumGC":         float64(ms.NumGC),         // количество GC
		"NumForcedGC":   float64(ms.NumForcedGC),   // количество принудительных GC

		// ===== Stack =====

		"StackInuse": float64(ms.StackInuse), // используемые байты stack
		"StackSys":   float64(ms.StackSys),   // байты под stack от ОС

		// ===== MCache / MSpan =====

		"MCacheInuse": float64(ms.MCacheInuse), // используемые MCache
		"MCacheSys":   float64(ms.MCacheSys),   // MCache от ОС
		"MSpanInuse":  float64(ms.MSpanInuse),  // используемые MSpan
		"MSpanSys":    float64(ms.MSpanSys),    // MSpan от ОС

		// ===== Прочее =====

		"BuckHashSys": float64(ms.BuckHashSys), // байты под hash buckets
		"OtherSys":    float64(ms.OtherSys),    // прочие аллокации рантайма

		// ===== Random Value =====

		"RandomValue": rand.Float64(),
	}
}

func CollectCustomMetrics() CustomMetrics {
	return CustomMetrics{
		"PollCount": atomic.AddInt64(&pollCount, 1),
	}
}

func (mc *MetricsClient) SendGauge(name string, value float64) {
	// val := strconv.FormatFloat(value, 'f', -1, 64)
	val := strconv.FormatFloat(value, 'g', -1, 64)

	uri := fmt.Sprintf("%s/update/gauge/%s/%s", mc.baseURL, url.PathEscape(name), val)
	mc.post(uri)
}

func (mc *MetricsClient) SendCounter(name string, value int64) {
	uri := fmt.Sprintf("%s/update/counter/%s/%d", mc.baseURL, name, value)
	mc.post(uri)
}

func (mc *MetricsClient) post(uri string) {
	resp, err := mc.client.R().Post(uri)
	if err != nil {
		log.Printf("send failed: %v", err)
		return
	}

	log.Printf("[DEBUG] POST %s -> %s", uri, resp.Status())
}

type NetAddress struct {
	Host string
	Port int
}

func (a *NetAddress) String() string {
	return fmt.Sprintf("http://%s:%d", a.Host, a.Port)
}

func (a *NetAddress) Set(value string) error {
	host, portStr, err := net.SplitHostPort(value)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	a.Host = host
	a.Port = port

	if a.Host == "" {
		a.Host = "localhost"
	}
	return nil
}

func main() {
	a := &NetAddress{
		Host: "localhost",
		Port: 8080,
	}

	flag.Var(a, "a", "Sending to server metrics http://host:port")
	flag.Parse()
	if a.Host == "" {
		a.Host = "localhost"
	}

	if a.Port == 0 {
		log.Fatal("invalid port")
	}

	client := NewMetricsClient(a.String())

	if *pollInterval <= 0 {
		log.Fatal("poll-interval must be > 0")
	}
	if *reportInterval <= 0 {
		log.Fatal("report-interval must be > 0")
	}

	pollTicker := time.NewTicker(*pollInterval)
	reportTicker := time.NewTicker(*reportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	for {
		select {

		case <-pollTicker.C:
			runtimeMetricsGauge.Store(CollectRuntimeMetrics())
			runtimeMetricsCounter.Store(CollectCustomMetrics())

		case <-reportTicker.C:

			if m, ok := runtimeMetricsGauge.Load().(RuntimeMetrics); ok {
				for name, value := range m {
					client.SendGauge(name, value)
				}
			}

			if m, ok := runtimeMetricsCounter.Load().(CustomMetrics); ok {
				for name, value := range m {
					client.SendCounter(name, value)
				}
			}

		}

	}
}
