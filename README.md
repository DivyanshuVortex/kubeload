# Kubeload 🚀

**Kubeload** is a lightweight CLI tool to stress-test your Kubernetes cluster. It acts as a stopwatch to measure exactly how fast your cluster can scale applications up and down.

## Quick Start

### 1. Build
```bash
go build -o kubeload main.go
```

### 2. Run
Make sure you are connected to a Kubernetes cluster (e.g., `kind`, `minikube`), then run a load test:

```bash
# Basic test (1 app, 3 cycles, 0 to 3 replicas)
./kubeload run

# Heavy test (10 parallel apps, 5 cycles, 0 to 5 replicas, 5 second wait)
./kubeload run --workers 10 --cycles 5 --min-replicas 0 --max-replicas 5 --dwell 5s
```

*Run `./kubeload run --help` to see all configuration flags.*

## How it Works

Kubeload spins up multiple concurrent "workers". Each worker creates a Kubernetes Deployment and rapidly scales it up and down, measuring the exact time it takes for the pods to fully start and completely terminate.

```mermaid
flowchart LR
    CLI[Kubeload] --> |1. Creates Deployments| K8s[K8s API]
    K8s --> |2. Scale Up (Time it!)| Pods
    K8s --> |3. Scale Down (Time it!)| Pods
    Pods --> |4. Report Latency| CLI
```

## Understanding the Results

When the test finishes, it automatically cleans up the Deployments and prints a summary table:

```text
=== Results: 10 worker(s) × 5 cycle(s) ===
Direction   Count   Min       Mean      P50       P95       P99       Max       Errors
--------------------------------------------------------------------------
up          50      2.037s    4.479s    4.396s    7.533s    7.933s    7.933s    0
down        50      1.02s     2.279s    2.124s    3.988s    4.27s     4.27s     0
```

- **Direction (`up` / `down`)**: Were pods being created or deleted?
- **Mean / Max**: The average and longest times it took to scale.
- **P95 / P99**: Percentiles (e.g., a P95 of 7.5s means 95% of events finished in under 7.5 seconds).
- **Errors**: Number of scaling operations that timed out.
