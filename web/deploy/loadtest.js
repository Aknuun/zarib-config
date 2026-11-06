import http from "k6/http";
import { check, sleep } from "k6";

const base = __ENV.BASE_URL || "https://zarib.app11.ir";
const sub = __ENV.SUB_URL || "";
const vus = Number(__ENV.VUS || 5);
const duration = __ENV.DURATION || "30s";

export const options = {
  vus,
  duration,
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<4000"],
  },
};

export default function () {
  const health = http.get(`${base}/healthz`);
  check(health, { "health 200": (r) => r.status === 200 });

  if (sub) {
    const inspect = http.post(
      `${base}/api/inspect`,
      JSON.stringify({ sub_url: sub }),
      { headers: { "Content-Type": "application/json" }, timeout: "60s" },
    );
    check(inspect, {
      "inspect not 5xx": (r) => r.status < 500,
      "inspect answered": (r) => r.body && r.body.length > 0,
    });
  }

  sleep(1);
}
