# Load Test Results

Testing performed with `hey` (HTTP load generator), against `/api/search` (limit: 20 capacity, refill 5 tokens/sec).

## Single gateway instance baseline
Command: `hey -z 5s -c 10 http://localhost:8080/api/search`

- Total requests: 114,224 (~22,841 req/sec)
- Allowed (200): 44
- Rejected (429): 114,180

Matches theoretical expectation: ~20 capacity + ~25 refilled over 5 sec ≈ 45 allowed.

## Distributed correctness — two concurrent gateway instances
Two gateway instances (`:8080` and `:8081`) load tested **simultaneously** (backgrounded, same start time), both hitting the same Redis-backed rate limiter for `/api/search`.

Command:
`hey -z 5s -c 10 http://localhost:8080/api/search > /tmp/gw1.txt &
hey -z 5s -c 10 http://localhost:8081/api/search > /tmp/gw2.txt &
wait`
Results:
| Instance | Total requests | Allowed (200) | Rejected (429) |
|----------|----------------|----------------|------------------|
| :8080    | 102,030        | 16             | 102,014          |
| :8081    | 102,973        | 28             | 102,945          |
| **Combined** | **205,003** | **44**         | **204,959**      |

## Conclusion
Combined allowed requests across both instances (44) matches the single-instance baseline (44) — the limit did **not** double when a second gateway instance was added. This confirms the rate limit state is correctly shared via Redis across distributed gateway nodes, holding under concurrent load exceeding 20k req/sec per instance.
