# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**


## Запуск pprof

# 

```
 pprof -top -diff_base=profiles/base.pprof_first profiles/base.pprof
File: server
Type: inuse_space
Time: 2026-04-19 15:58:29 MSK
Showing nodes accounting for -176.55MB, 90.82% of 194.39MB total
Dropped 50 nodes (cum <= 0.97MB)
      flat  flat%   sum%        cum   cum%
  -90.79MB 46.70% 46.70%  -105.75MB 54.40%  compress/flate.NewWriter (inline)
  -63.80MB 32.82% 79.53%   -63.80MB 32.82%  runtime.mallocgc
  -14.96MB  7.70% 87.22%   -14.96MB  7.70%  compress/flate.(*compressor).initDeflate (inline)
      -3MB  1.54% 88.77%       -3MB  1.54%  internal/sync.runtime_SemacquireMutex
   -2.50MB  1.29% 90.05%    -2.50MB  1.29%  encoding/json.(*decodeState).literalStore
      -2MB  1.03% 91.08%  -110.75MB 56.97%  github.com/ilyinon/go-musthave-metrics/internal/handler.(*UpdateJSONHandler).ServeHTTP
    0.50MB  0.26% 90.82%     1.50MB  0.77%  net/http.ListenAndServe (inline)
```