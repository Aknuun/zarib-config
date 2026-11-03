import type { InspectResponse, Job, LiveStats } from "./types";

export class ApiError extends Error {
  code: string;
  status: number;
  log: string[];

  constructor(message: string, code: string, status = 0, log: string[] = []) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
    this.log = log;
  }
}

interface RequestOptions {
  timeout?: number;
  retries?: number;
}

export async function request<T>(
  path: string,
  init: RequestInit = {},
  { timeout = 25_000, retries = 0 }: RequestOptions = {},
): Promise<T> {
  let lastError: unknown;
  for (let attempt = 0; attempt <= retries; attempt += 1) {
    const controller = new AbortController();
    const timer = window.setTimeout(() => controller.abort(), timeout);
    try {
      const response = await fetch(path, {
        ...init,
        headers: { "Content-Type": "application/json", ...(init.headers || {}) },
        signal: controller.signal,
      });
      window.clearTimeout(timer);
      const data = await response.json().catch(() => null);
      if (!response.ok) {
        throw new ApiError(
          (data && data.error) || response.statusText,
          (data && data.code) || (response.status >= 500 ? "network" : "unknown"),
          response.status,
          (data && data.log) || [],
        );
      }
      if (data === null) {
        throw new ApiError("Invalid server response.", "unknown", response.status);
      }
      return data as T;
    } catch (error) {
      window.clearTimeout(timer);
      const aborted = error instanceof DOMException && error.name === "AbortError";
      const apiError =
        error instanceof ApiError
          ? error
          : new ApiError(
              aborted ? "Request timed out." : "Could not reach the server.",
              aborted ? "timeout" : "network",
            );
      lastError = apiError;
      if (attempt < retries && (apiError.code === "network" || apiError.code === "timeout")) {
        await new Promise((resolve) => window.setTimeout(resolve, 600 * (attempt + 1)));
        continue;
      }
      throw apiError;
    }
  }
  throw lastError;
}

export function inspectSubscription(subUrl: string) {
  return request<InspectResponse>(
    "/api/inspect",
    { method: "POST", body: JSON.stringify({ sub_url: subUrl }) },
    { retries: 1, timeout: 30_000 },
  );
}

export function createJob(input: {
  sub_url: string;
  pick: number;
  size_mb: number;
  sync: boolean;
}) {
  return request<{ id: string; status: string }>("/api/jobs", {
    method: "POST",
    body: JSON.stringify({
      sub_url: input.sub_url,
      pick: input.pick,
      size_mb: input.size_mb,
      wait: 2,
      drift: 0,
      sync: input.sync,
      core: "auto",
    }),
  });
}

export function createBatchJob(input: {
  sub_url: string;
  size_mb: number;
  sync: boolean;
  picks: number[];
}) {
  return request<{ id: string; status: string }>("/api/jobs", {
    method: "POST",
    body: JSON.stringify({
      sub_url: input.sub_url,
      mode: "batch",
      picks: input.picks,
      size_mb: input.size_mb,
      wait: 2,
      drift: 0,
      sync: input.sync,
      core: "auto",
    }),
  });
}

export function getJob(id: string) {
  return request<Job>(`/api/jobs/${id}`, {}, { timeout: 120_000 });
}

export function cancelJob(id: string) {
  return request<{ ok: boolean }>(`/api/jobs/${id}/cancel`, { method: "POST" });
}

export function pingTarget(target: string) {
  return request<{ ok: boolean; latency_ms: number }>(
    `/api/ping?target=${encodeURIComponent(target)}`,
  );
}

export function geoTarget(target: string) {
  return request<{ country: string; tunnel: boolean }>(
    `/api/geo?target=${encodeURIComponent(target)}`,
  );
}

export function sendPresence(id: string) {
  return request<LiveStats>("/api/presence", {
    method: "POST",
    body: JSON.stringify({ id }),
  });
}

export function sendReport(input: {
  description: string;
  sub_url: string;
  result: unknown;
  meta?: Record<string, string>;
  view?: string;
  error_code?: string;
  error_message?: string;
  error_log?: string[];
}) {
  return request<{ ok: boolean }>("/api/report", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function getHealth() {
  return request<{ ok: boolean; version: string }>("/healthz");
}
