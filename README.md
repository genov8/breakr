# breakr - Circuit Breaker for Go

`breakr` is a lightweight, production-ready Circuit Breaker implementation for Go.  
It helps protect your services from cascading failures and ensures high availability.

## Why Breakr?

Breakr is a modern, observability-first circuit breaker for Go.

While many libraries focus on complex configuration, Breakr focuses on:

- Simplicity
- Production readiness
- Built-in Prometheus integration
- Clean and minimal API

Breakr is designed for modern microservices where monitoring and observability matter.

## 📦 Installation

```sh
go get github.com/genov8/breakr
```

---

## Usage
### Simple Example
```go
package main

import (
	"fmt"
	"github.com/genov8/breakr"
	"time"
)

func main() {
	cb := breakr.New(breakr.Config{
		FailureThreshold: 3,
		ResetTimeout:     5 * time.Second,
		ExecutionTimeout: 2 * time.Second,
	})

	for i := 0; i < 10; i++ {
		result, err := cb.Execute(func() (interface{}, error) {
			return nil, fmt.Errorf("error")
		})

		if err != nil {
			fmt.Printf("[Request %d] Circuit Breaker blocked: %v\n", i, err)
		} else {
			fmt.Printf("[Request %d] Success: %v\n", i, result)
		}

		time.Sleep(500 * time.Millisecond)
	}
}
```

## ⚙️ Configuration

| Parameter | Description |
| --- | --- |
| FailureThreshold | Number of consecutive failures before CB enters Open state.|
| ResetTimeout | Time before CB moves to Half-Open. |
| ExecutionTimeout | Maximum execution time for a protected function. |
| WindowSize | Duration of sliding time window (e.g., `2s`). Only failures within this window are counted toward the threshold. Use `0` to disable. |
| FailureCodes | List of HTTP status codes considered failures (e.g., `[500, 502, 503]`). **If omitted, all errors trigger the breaker.** |

## 📊 Metrics (Prometheus)

`breakr` can export **Prometheus metrics** to help you observe circuit breaker behavior in production.

Metrics are **optional** and enabled only if you provide a metrics instance.

### Enable metrics

```go
import (
    "github.com/genov8/breakr"
    "github.com/genov8/breakr/config"
    "github.com/genov8/breakr/metrics"
)

m := metrics.NewMetrics("breakr")

cb := breakr.New(config.Config{
    FailureThreshold: 3,
    ResetTimeout:     5 * time.Second,
    ExecutionTimeout: 2 * time.Second,
    Metrics:          m,
})
```

### Exporting metrics endpoint

`breakr` does not start an HTTP server by itself.

To expose metrics, register the Prometheus handler in your application:

```go
import (
    "net/http"
    
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

http.Handle("/metrics", promhttp.Handler())
http.ListenAndServe(":2112", nil)
```
### Exported metrics

| Metric | Type | Description |
|------|------|-------------|
| `breakr_requests_total` | Counter | Total number of requests by status and state |
| `breakr_execution_duration_seconds` | Histogram | Execution duration of protected calls |
| `breakr_state` | Gauge | Current circuit breaker state |
| `breakr_state_transitions_total` | Counter | Number of state transitions |

#### Labels

- `status`: `success`, `error`, `timeout`, `blocked`, `ignored_error`
- `state`: `Closed`, `Open`, `HalfOpen`

### Visualization

`breakr` only exports Prometheus metrics.

You can visualize them using **Grafana** or any other Prometheus-compatible monitoring tool.
Dashboard configuration is left to the user.

---

You can configure breakr using JSON or YAML instead of manual setup.

#### 📝 JSON Example
```json
{
  "failure_threshold": 3,
  "reset_timeout": "5s",
  "execution_timeout": "2s",
  "window_size": "10s",
  "failure_codes": [500, 502, 503]
}
```
```go
config, err := config.LoadConfigJSON("config.json")
if err != nil {
    log.Fatalf("Error loading config: %v", err)
}
cb := breakr.New(*config)
```
#### 📝 YAML Example
```yaml
failure_threshold: 3
reset_timeout: "5s"
execution_timeout: "2s"
window_size: "10s"
failure_codes:
  - 500
  - 502
  - 503
```
```go
config, err := config.LoadConfigYAML("config.yaml")
if err != nil {
    log.Fatalf("Error loading config: %v", err)
}
cb := breakr.New(*config)
```

### 🌐 Example 1: Circuit Breaker with HTTP Requests

This example demonstrates how `Breakr` can handle **unstable HTTP requests**.  
The circuit breaker will **trip after 3 failed requests** and block further requests until `ResetTimeout` expires.

```go
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/genov8/breakr"
)

func unstableHandler(w http.ResponseWriter, r *http.Request) {
	if rand.Float32() < 0.7 {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	// Start an unstable test server
	http.HandleFunc("/test", unstableHandler)
	go func() {
		log.Println("Starting test server on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// Configure Circuit Breaker
	cb := breakr.New(breakr.Config{
		FailureThreshold: 3,
		ResetTimeout:     5 * time.Second,
		ExecutionTimeout: 2 * time.Second,
	})

	// Send requests and observe how CB behaves
	for i := 1; i <= 20; i++ {
		result, err := cb.Execute(func() (interface{}, error) {
			resp, err := http.Get("http://localhost:8080/test")
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 500 {
				return nil, fmt.Errorf("server error: %d", resp.StatusCode)
			}
			return "Request successful", nil
		})

		if err != nil {
			fmt.Printf("[Request %d] Circuit Breaker blocked: %v\n", i, err)
		} else {
			fmt.Printf("[Request %d] Success: %v\n", i, result)
		}

		time.Sleep(500 * time.Millisecond)
	}
}
```
### ⚡ Example 2: Filtering Specific Failure Codes
This example shows how to configure Breakr to only react to certain HTTP error codes (500, 502, 503) while ignoring others like 404.
```go
package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/genov8/breakr"
)

// APIError — custom error type with an HTTP status code
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) Code() int {
	return e.StatusCode
}

func main() {
	// Create a Circuit Breaker with FailureCodes support
	cb := breakr.New(breakr.Config{
		FailureThreshold: 3,
		ResetTimeout:     5 * time.Second,
		ExecutionTimeout: 2 * time.Second,
		FailureCodes:     []int{500, 502, 503}, // Reacts only to these error codes
	})

	// API call returning 500 (critical failure)
	fail500 := func() (interface{}, error) {
		return nil, &APIError{StatusCode: 500, Message: "Internal Server Error"}
	}

	// API call returning 404 (ignored error)
	fail404 := func() (interface{}, error) {
		return nil, &APIError{StatusCode: 404, Message: "Not Found"} // Ignored
	}

	// Successful API call
	success := func() (interface{}, error) {
		return "Success", nil
	}

	// Simulating requests
	for i := 1; i <= 5; i++ {
		result, err := cb.Execute(fail404)
		if err != nil {
			fmt.Printf("[Request %d] Ignored Error: %v\n", i, err)
		} else {
			fmt.Printf("[Request %d] Success: %v\n", i, result)
		}
	}

	for i := 6; i <= 10; i++ {
		result, err := cb.Execute(fail500)
		if err != nil {
			fmt.Printf("[Request %d] Failure: %v\n", i, err)
		} else {
			fmt.Printf("[Request %d] Success: %v\n", i, result)
		}
	}

	// Circuit Breaker should now be in Open state
	for i := 11; i <= 13; i++ {
		result, err := cb.Execute(fail500)
		if err != nil {
			fmt.Printf("[Request %d] Circuit Breaker blocked: %v\n", i, err)
		} else {
			fmt.Printf("[Request %d] Success: %v\n", i, result)
		}
	}

	// Waiting for ResetTimeout
	fmt.Println("⏳ Waiting for ResetTimeout...")
	time.Sleep(6 * time.Second)

	// Checking Half-Open state, should allow a successful request
	result, err := cb.Execute(success)
	if err != nil {
		fmt.Printf("[Request 14] Failure: %v\n", err)
	} else {
		fmt.Printf("[Request 14] Success: %v\n", result)
	}

	// Should fully recover after success
	result, err = cb.Execute(success)
	if err != nil {
		fmt.Printf("[Request 15] Failure: %v\n", err)
	} else {
		fmt.Printf("[Request 15] Success: %v\n", result)
	}
}
```
### 🌐 Example 3: Sliding window
This example demonstrates how Breakr uses a sliding time window to only consider recent failures toward the threshold.
```go
package main

import (
	"fmt"
	"github.com/genov8/breakr"
	"github.com/genov8/breakr/config"
	"time"
)

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) Code() int {
	return e.StatusCode
}

func main() {
	cb := breakr.New(config.Config{
		FailureThreshold: 2,
		ResetTimeout:     3 * time.Second,
		ExecutionTimeout: 1 * time.Second,
		WindowSize:       2 * time.Second,
		FailureCodes:     []int{500},
	})

	fail := func() (interface{}, error) {
		return nil, &APIError{StatusCode: 500, Message: "Internal Server Error"}
	}

	success := func() (interface{}, error) {
		return "OK", nil
	}

	_ = success

	cb.Execute(fail)
	fmt.Println("[1] First failure")

	time.Sleep(3 * time.Second)

	cb.Execute(fail)
	fmt.Println("[2] Second failure (alone in window)")

	fmt.Printf("[2] State: %s\n", cb.State())

	cb.Execute(fail)
	fmt.Println("[3] Third failure — should trip breaker")

	fmt.Printf("[3] State: %s\n", cb.State())

	// Output:
	// [1] First failure
	// [2] Second failure (alone in window)
	// [2] State: Closed
	// [3] Third failure — should trip breaker
	// [3] State: Open
}

```
### 🧪 Example 4: Execute with context
This example shows how to use `ExecuteCtx` to control execution timeout via `context.Context`.

```go
ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
defer cancel()

result, err := cb.ExecuteCtx(ctx, func(ctx context.Context) (interface{}, error) {
    select {
    case <-time.After(1 * time.Second):
        return "done", nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
})

if err != nil {
    fmt.Printf("⛔ Request failed: %v\n", err)
} else {
    fmt.Printf("✅ Result: %v\n", result)
}
```

## 📜 Circuit Breaker States

- Closed → Everything works fine, requests are allowed.
- Open → Requests are blocked after reaching the failure threshold.
- Half-Open → A test request is allowed to check if recovery is possible.

``` 
[Closed] → (errors > threshold) → [Open] → (timeout expires) → [Half-Open]
       ↑                                            ↓
       └── (success) ← (failure) ───────────────────┘
```

## 🎯 Key Features

- [x] Protects against cascading failures
- [x] Limits retries to avoid overloading a failing service
- [x] Fast & lightweight
- [x] Supports execution timeouts
- [x] Allows filtering which errors trigger the breaker (`FailureCodes`)
- [x] JSON & YAML configuration support
- [x] Sliding window strategy — count only recent failures in a time window
- [x] Execute with `context.Context` via `ExecuteCtx`
- [x] Optional Prometheus metrics for observability
