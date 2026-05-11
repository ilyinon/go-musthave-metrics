// Package noosexitmain provides an analyzer that forbids
// direct usage of os.Exit inside main.main functions.
// Command staticlint runs custom multichecker.
//
// Usage:
//
//	go run ./cmd/staticlint ./...
//
// or build:
//
//	go build -o staticlint ./cmd/staticlint
//	./staticlint ./...
//
// Included analyzers:
//
// 1. Standard analyzers:
//   - assign: useless assignments
//   - bools: boolean simplifications
//   - printf: format issues
//   - shadow: variable shadowing
//
// 2. Staticcheck (SA):
//   - detects bugs, performance issues, misuse
//
// 3. Additional staticcheck analyzers:
//   - simple: code simplifications
//   - stylecheck: style issues
//
// 4. Extra analyzers:
//   - atomic: atomic misuse
//   - copylock: lock copying issues
//
// 5. Custom analyzer:
//   - noosexit: запрещает os.Exit в main.main
package noosexitmain
