// TalkEx messaging engine load test — hits POST /conversations/send
// with a burst pattern that ramps to 1000 req/sec and holds for 2 min.
//
// Usage:
//   TALKEX_BASE_URL=https://api.talkex.io TALKEX_API_KEY=sk_... k6 run k6-messaging.js
//
// The target endpoint is designed to be sub-200ms P95 from the Mumbai
// region — anything above that means the messaging queue or the DB
// primary is under pressure.
import http from 'k6/http'
import { check, sleep } from 'k6'
import { Counter, Trend } from 'k6/metrics'

const BASE = __ENV.TALKEX_BASE_URL || 'http://localhost:8080'
const KEY = __ENV.TALKEX_API_KEY || ''

if (!KEY) throw new Error('TALKEX_API_KEY is required')

// Per-tenant target contact — replace with a real ID from your workspace.
const CONTACT_ID = __ENV.CONTACT_ID || 'ct_loadtest_replaceme'

const rate429 = new Counter('rate_limit_hits')
const enqueueLatency = new Trend('enqueue_latency_ms', true)

export const options = {
  scenarios: {
    burst: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 500,
      stages: [
        { duration: '30s', target: 100 },   // warm-up
        { duration: '30s', target: 500 },   // sustained
        { duration: '2m',  target: 1000 },  // burst — 1000 rps for 2 min
        { duration: '30s', target: 0 },     // cool-down
      ],
    },
  },
  thresholds: {
    'http_req_duration{scenario:burst}': ['p(95)<200', 'p(99)<500'],
    'http_req_failed':                    ['rate<0.02'],   // <2 % failures
    'enqueue_latency_ms':                 ['p(95)<200'],
  },
}

export default function () {
  const idempotencyKey = `loadtest-${__VU}-${__ITER}-${Date.now()}`
  const body = JSON.stringify({
    contact_id: CONTACT_ID,
    channel: 'talkex',
    body: `loadtest msg ${idempotencyKey}`,
  })

  const res = http.post(`${BASE}/conversations/send`, body, {
    headers: {
      Authorization: `Bearer ${KEY}`,
      'Content-Type': 'application/json',
      'Idempotency-Key': idempotencyKey,
    },
    tags: { scenario: 'burst' },
  })

  if (res.status === 429) rate429.add(1)
  enqueueLatency.add(res.timings.duration)

  check(res, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
    'has message_id': (r) => r.json('message_id') !== '',
  })

  // Bake a tiny think-time so VUs don't spin
  sleep(0.05)
}
