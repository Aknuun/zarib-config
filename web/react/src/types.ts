export interface Usage {
  upload: number;
  download: number;
  used: number;
  total: number | null;
  remaining: number | null;
  expire: number | null;
  days_left: number | null;
}

export interface SubConfig {
  index: number;
  protocol: string;
  target: string;
  remark: string;
  claimed: number | null;
  uri: string;
  pingMs?: number | null;
  tunnel?: boolean | null;
  geoCountry?: string;
}

export interface InspectResponse {
  usage: Usage | null;
  configs: SubConfig[];
  count: number;
  insecure: boolean;
}

export interface SubState extends InspectResponse {
  url: string;
  host: string;
}

export type JobStatus = "queued" | "running" | "done" | "error" | "cancelled";

export interface RatioResult {
  version: string;
  tested_at: string;
  config: string | null;
  configs_tested?: number;
  core: string;
  xray_version: string;
  downloaded_bytes: number;
  usage_before: Usage;
  usage_after: Usage;
  delta_upload: number;
  delta_download: number;
  delta_total: number;
  ratio: number;
  ratio_raw: number;
  background_bytes_per_second: number;
  background_bytes_in_test: number;
  claimed_ratio: number | null;
  claimed_matches: boolean | null;
  verdict: string | null;
  remaining_bytes: number | null;
  total_bytes: number | null;
  upload_after: number | null;
  download_after: number | null;
  duration_seconds: number;
}

export interface Job {
  id: string;
  status: JobStatus;
  stage: string;
  progress: number;
  message: string;
  elapsed: number;
  created: number;
  result: RatioResult | null;
  error: string;
  error_code: string;
  log: string[];
  mode?: "single" | "batch";
  pick: number;
  size_mb: number;
  core: string;
}

export interface FailStats {
  total?: number;
  connect?: number;
  fetch?: number;
  core?: number;
  timeout?: number;
  usage?: number;
  other?: number;
}

export interface LiveBucket {
  date?: string;
  month?: string;
  tests: number;
  with_ratio: number;
  normal: number;
  invalid: number;
  tunnel: number;
  non_tunnel: number;
  unknown_tunnel: number;
  downloaded: number;
  deducted: number;
  unique_ips?: number;
  proxy: number;
  claimed_tests?: number;
  claimed_avg?: number;
  claimed_real_avg?: number;
  claimed_diff_avg?: number;
  claimed_diff_pct?: number;
  failed: FailStats;
}

export interface LiveStats {
  online: number;
  h24: LiveBucket;
  d7: LiveBucket;
  d30: LiveBucket;
  overall: LiveBucket;
  today?: LiveBucket;
  monthly?: LiveBucket;
}

export interface JobRecord {
  id: string;
  url: string;
  host: string;
  pick: number;
  label: string;
  startedAt: number;
  size: number;
  bg: boolean;
  total: number;
  completed: number;
  batch?: boolean;
  configs: Array<Pick<SubConfig, "index" | "protocol" | "target" | "remark">>;
  queue: number[];
  results: Array<{
    pick: number;
    result: RatioResult | null;
    errorCode?: string;
    errorMessage?: string;
    errorLog?: string[];
  }>;
}

export interface TestRow {
  pick: number;
  config: Pick<SubConfig, "index" | "protocol" | "target" | "remark">;
  result: RatioResult | null;
  errorCode?: string;
  errorMessage?: string;
  errorLog?: string[];
}

export interface Aggregate {
  ratio: number;
  downloaded: number;
  deducted: number;
  background: number;
  excluded: number;
  count: number;
  duration: number;
  testedAt: string;
  core: string;
  xrayVersion: string;
  claimed: number | null;
  nonRound: boolean;
}
