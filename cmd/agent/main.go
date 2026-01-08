package main

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
)

var pollCount int64

type RuntimeMetrics map[string]float64
type CustomMetrics map[string]int

type MetricsClient struct {
	baseURL string
	client  *resty.Client
}

func NewMetricsClient(baseURL string) *MetricsClient {
	client := resty.New().
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
	counter := atomic.AddInt64(&pollCount, 1)

	// Заполняем map метрик
	return CustomMetrics{
		"PollCount": int(counter),
	}
}

type Number interface {
	~int64 | ~float64
}

func (mc *MetricsClient) SendGauge(name string, value float64) {
	send(mc, "gauge", name, value)
}

func (mc *MetricsClient) SendCounter(name string, value int64) {
	send(mc, "counter", name, value)
}

func send[T Number](mc *MetricsClient, t, name string, value T) {
	uri := fmt.Sprintf("%s/update/%s/%s/%v", mc.baseURL, t, name, value)

	resp, err := mc.client.R().Post(uri)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Status())
}

func main() {
	client := NewMetricsClient("http://localhost:8080")

	var pollInterval int = 2
	var reportInterval int = 10

	for {
		time.Sleep(time.Duration(pollInterval) * time.Second)

		// metrics := CollectRuntimeMetrics()

		time.Sleep(time.Duration(reportInterval-pollInterval) * time.Second)

		for name, value := range CollectRuntimeMetrics() {
			// Отправить в /update/gauge/{name}/{value}
			client.SendGauge(name, value)
		}
		for name, value := range CollectCustomMetrics() {
			// Отправить в /update/counter/{name}/{value}
			client.SendCounter(name, int64(value))
		}
	}

}
