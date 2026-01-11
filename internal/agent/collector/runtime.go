package collector

import (
	"math/rand"
	"runtime"

	"github.com/ilyinon/go-musthave-metrics/internal/model"
)

func Runtime() model.RuntimeMetrics {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return model.RuntimeMetrics{
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
