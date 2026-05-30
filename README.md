# Kubeload

**Kubeload** is a Kubernetes scale latency load generator. It is a lightweight CLI tool designed to stress-test your Kubernetes cluster and benchmark its scaling performance. It measures how long the control plane, scheduler, and node autoscaler take to actually spin up and tear down pods during rapid scaling events.

## How it Works (Architecture)

Kubeload simulates load by creating concurrent "workers", where each worker manages its own Kubernetes Deployment and rapidly scales it up and down in a loop.

```mermaid
flowchart TD
    CLI[CLI Entrypoint (Cobra)] --> |Initializes K8s Client| RunCmd[Run Command]
    RunCmd --> |Creates Deployments| K8s[Kubernetes API Server]
    
    subgraph Workers [Concurrent Workers (Goroutines)]
        W1[Worker 0]
        W2[Worker 1]
        WN[Worker N]
    end
    
    RunCmd --> Workers
    
    W1 --> |Scale to Max Replicas| K8s
    W1 -.-> |Measure Scale-Up Latency| Metrics[Metrics Aggregator]
    W1 --> |Scale to Min Replicas| K8s
    W1 -.-> |Measure Scale-Down Latency| Metrics
    
    Metrics --> |Generate Report| Output[Console Output]
```

1. **Initialization:** Kubeload authenticates with your current Kubernetes context.
2. **Setup:** It creates distinct Deployments (e.g., `kubeload-0`, `kubeload-1`) depending on the configured number of concurrent workers.
3. **Execution Cycles:** For each configured cycle, every worker will:
   - Scale its deployment up to `max-replicas`.
   - Measure the exact time it takes for all pods to reach a `Ready` state.
   - Wait for a specified `dwell` time.
   - Scale its deployment down to `min-replicas`.
   - Measure the exact time it takes for pods to fully terminate.
4. **Reporting:** Once all cycles are complete across all workers, a consolidated latency report is printed to the console.
5. **Cleanup:** Test deployments are automatically deleted.

## Installation

Assuming you have Go installed, you can build the project from source:

```bash
go build -o kubeload main.go
```

## Usage

```bash
./kubeload run [flags]
```

### Examples

**Run a basic load test:**
(1 worker, 3 cycles, scaling from 0 to 3 replicas)
```bash
./kubeload run
```

**Run a heavy concurrent load test:**
(10 concurrent workers, 5 cycles each, scaling from 0 to 5 replicas, with a 5-second dwell time)
```bash
./kubeload run --workers 10 --cycles 5 --min-replicas 0 --max-replicas 5 --dwell 5s
```

### Flags

- `--name` (default: `kubeload`): Base name for deployments.
- `--namespace` (default: `default`): Kubernetes namespace to run tests in.
- `--image` (default: `nginx:alpine`): Container image used for test pods.
- `--workers` (default: `1`): Number of concurrent workers (each gets its own deployment).
- `--cycles` (default: `3`): Number of scale up/down cycles per worker.
- `--min-replicas` (default: `0`): Replica count to scale down to.
- `--max-replicas` (default: `3`): Replica count to scale up to.
- `--dwell` (default: `2s`): Hold time at max-replicas before scaling down.
- `--timeout` (default: `10m`): Global run timeout.
- `--cleanup` (default: `true`): Delete deployments after the run completes.

## Understanding the Results

When the test finishes, `kubeload` will print a table summarizing the latencies for scaling your deployments up and down.

**Example Output:**
```text
=== Results: 10 worker(s) × 5 cycle(s) ===
Direction   Count   Min       Mean      P50       P95       P99       Max       Errors
--------------------------------------------------------------------------
up          50      2.037s    4.479s    4.396s    7.533s    7.933s    7.933s    0
down        50      1.02s     2.279s    2.124s    3.988s    4.27s     4.27s     0
```

- **Direction (`up` or `down`)**: Denotes whether pods were being created or terminated.
- **Count**: Total number of samples collected (Workers × Cycles).
- **Min / Max / Mean**: The shortest, longest, and average times taken for the pods to fully start (reach a Ready state) or fully terminate. 
- **P50 / P95 / P99**: Percentiles. For example, a P95 `up` latency of 7.5s means that 95% of your scale-up events completed within 7.5 seconds.
- **Errors**: Number of scaling operations that timed out or failed to complete.

