import http from "k6/http";
import { check, sleep } from "k6";
import { Trend, Rate } from "k6/metrics";

const baseUrl = __ENV.BASE_URL || "http://localhost:8080";

const profileUpdateLatency = new Trend("profile_update_latency", true);
const profileUpdateErrors = new Rate("profile_update_errors");

const WARMUP_VUS = Number(__ENV.WARMUP_VUS || 5);
const STEADY_VUS = Number(__ENV.STEADY_VUS || 30);
const RAMPDOWN_VUS = Number(__ENV.RAMPDOWN_VUS || 0);
const WARMUP_DURATION = __ENV.WARMUP_DURATION || "30s";
const STEADY_DURATION = __ENV.STEADY_DURATION || "2m";
const RAMPDOWN_DURATION = __ENV.RAMPDOWN_DURATION || "30s";

export const options = {
  stages: [
    { duration: WARMUP_DURATION, target: WARMUP_VUS },
    { duration: STEADY_DURATION, target: STEADY_VUS },
    { duration: RAMPDOWN_DURATION, target: RAMPDOWN_VUS }
  ],
  thresholds: {
    http_req_failed: ["rate<0.05"],
    http_req_duration: ["p(95)<600"],
    profile_update_errors: ["rate<0.05"]
  },
  summaryTrendStats: ["avg", "min", "med", "p(90)", "p(95)", "p(99)", "max"]
};

function randomEmail() {
  return `bench_${Date.now()}_${Math.floor(Math.random() * 100000)}@example.com`;
}

export function setup() {
  const payload = JSON.stringify({
    email: randomEmail(),
    username: `bench_user_${Math.floor(Math.random() * 1000000)}`,
    full_name: "Benchmark User",
    age_group: "18-24"
  });

  const res = http.post(`${baseUrl}/api/v1/users`, payload, {
    headers: { "Content-Type": "application/json" }
  });

  check(res, { "setup create user status 201": (r) => r.status === 201 });

  const body = res.json();
  return { userId: body.id };
}

export default function (data) {
  const payload = JSON.stringify({
    monthly_income: 3000 + Math.floor(Math.random() * 2000),
    monthly_expenses: 1500 + Math.floor(Math.random() * 1200),
    total_debt: 2000 + Math.floor(Math.random() * 15000),
    debt_apr: 12 + Math.random() * 20,
    savings_balance: 500 + Math.floor(Math.random() * 8000),
    credit_score: 580 + Math.floor(Math.random() * 180)
  });

  const res = http.put(`${baseUrl}/api/v1/users/${data.userId}/profile`, payload, {
    headers: { "Content-Type": "application/json" }
  });

  profileUpdateLatency.add(res.timings.duration);
  const ok = check(res, { "profile update status 200": (r) => r.status === 200 });
  profileUpdateErrors.add(!ok);

  sleep(Number(__ENV.THINK_TIME_SECONDS || 0.1));
}
