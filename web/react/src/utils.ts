export function formatBytes(value: number | null | undefined) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) return "—";
  let number = Number(value);
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  let index = 0;
  while (Math.abs(number) >= 1024 && index < units.length - 1) {
    number /= 1024;
    index += 1;
  }
  return `${number.toFixed(index === 0 ? 0 : 2)} ${units[index]}`;
}

export function isNoRatio(value: number) {
  return value >= 0.95 && value < 1.06;
}

export function isInvalidRatio(value: number) {
  return value > 0 && value < 0.95;
}

export function isNonRoundRatio(value: number) {
  if (!Number.isFinite(value) || value <= 0) return false;
  return Math.abs(value * 100 - Math.round(value * 100)) > 0.0001;
}

export function ratioNumber(value: number) {
  if (isNoRatio(value)) return "1.00";
  return value.toFixed(isNonRoundRatio(value) ? 3 : 2);
}

export function formatRatio(value: number | null | undefined) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) return "—";
  return `\u200e${ratioNumber(Number(value))}×`;
}

export function ratioPercent(value: number) {
  const min = 0.75;
  const max = 2.25;
  return Math.min(98, Math.max(2, ((value - min) / (max - min)) * 100));
}

export type RatioTone = "good" | "warn" | "danger";

export function ratioTone(value: number): RatioTone {
  if (isInvalidRatio(value)) return "danger";
  if (value >= 1.5) return "danger";
  if (value >= 1.06) return "warn";
  return "good";
}

export function formatDuration(seconds: number | null | undefined) {
  const total = Math.max(0, Math.round(Number(seconds) || 0));
  const mins = Math.floor(total / 60);
  const secs = total % 60;
  return `${String(mins).padStart(2, "0")}:${String(secs).padStart(2, "0")}`;
}

export function formatDate(value: string, lang: "fa" | "en") {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  try {
    return date.toLocaleString(lang === "fa" ? "fa-IR" : "en-GB", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return value;
  }
}

export function normalizeUrl(raw: string) {
  let value = raw.trim();
  if (!value) return "";
  if (!/^[a-z][a-z0-9+.-]*:\/\//i.test(value)) value = `https://${value}`;
  return value;
}

export function hostFromUrl(value: string) {
  try {
    return new URL(value).host || value;
  } catch {
    return value;
  }
}

export async function copyText(text: string) {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch {
    /* fall through */
  }
  try {
    const area = document.createElement("textarea");
    area.value = text;
    area.style.position = "fixed";
    area.style.opacity = "0";
    document.body.appendChild(area);
    area.select();
    const ok = document.execCommand("copy");
    area.remove();
    return ok;
  } catch {
    return false;
  }
}

export function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value));
}

export function errorKey(code: string | undefined) {
  const known = [
    "invalid_url",
    "blocked_url",
    "fetch_failed",
    "no_configs",
    "probe_failed",
    "resolve_failed",
    "xray_failed",
    "usage_unchanged",
    "timeout",
    "rate_limited",
    "busy",
    "duplicate_test",
    "subscription_expired",
    "subscription_limited",
    "subscription_broken",
    "subscription_unauthorized",
    "subscription_not_found",
    "subscription_rate_limited",
    "subscription_unavailable",
    "cancelled",
    "pick_out_of_range",
    "invalid_input",
    "not_found",
    "unknown",
    "network",
    "aborted",
  ];
  return known.includes(code || "") ? `err_${code}` : "err_unknown";
}

export function suggestionKey(code: string | undefined) {
  switch (code) {
    case "probe_failed":
    case "resolve_failed":
    case "timeout":
    case "xray_failed":
      return "errorSuggestDefault";
    case "fetch_failed":
    case "blocked_url":
    case "invalid_url":
    case "no_configs":
      return "errorSuggestLink";
    case "usage_unchanged":
      return "errorSuggestPanel";
    case "rate_limited":
    case "busy":
      return "errorSuggestWait";
    case "duplicate_test":
      return "errorSuggestDuplicate";
    case "subscription_expired":
    case "subscription_limited":
      return "errorSuggestSubscription";
    case "subscription_broken":
    case "subscription_unauthorized":
    case "subscription_not_found":
      return "errorSuggestLink";
    case "subscription_rate_limited":
    case "subscription_unavailable":
      return "errorSuggestWait";
    default:
      return "errorSuggestDefault";
  }
}

export const SELF_HOST_SCRIPT =
  'bash -c "$(curl -fsSL https://github.com/Aknuun/zarib-config/raw/refs/heads/main/run.sh)"';
