# Throttler  
[![Go Reference](https://pkg.go.dev/badge/github.com/gopal96685/throttler.svg)](https://pkg.go.dev/github.com/gopal96685/throttler)  
[![License](https://img.shields.io/github/license/gopal96685/throttler)](./LICENSE)  
[📖 Read the full article on Medium](https://medium.com/@gopal96685/introducing-throttler-a-go-library-for-simplified-task-management-and-rate-limiting-a1187510a57e)

**Throttler** is a Go library for managing task execution with features like:  
- **Rate Limiting** to handle API or database constraints.  
- **Task Scheduling** with retries, backoff, and priority management.  
- **Circuit Breaking** to avoid cascading failures.  

---

## Features  
- Flexible APIs for defining tasks: Use closures or the `Executable` interface.  
- Automatic retry with backoff for failed tasks.  
- Rate limit enforcement to handle throughput-heavy workloads.  
- Priority-based task execution with first-in-first-out (FIFO) fallback.  
- Built-in metrics collection for monitoring queue size and execution.  

---

## Installation  
```bash
go get github.com/gopal96685/throttler
