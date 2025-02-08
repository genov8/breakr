# breakr - Circuit Breaker for Go

`breakr` is a lightweight, production-ready Circuit Breaker implementation for Go.  
It helps protect your services from cascading failures and ensures high availability.

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

## 🌐 Example with HTTP Requests

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