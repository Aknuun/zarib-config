import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { TooltipProvider } from "@cloudflare/kumo";
import { Header } from "./components/Header";
import { Landing } from "./components/Landing";
import { SetupView } from "./components/SetupView";
import { ProgressView } from "./components/ProgressView";
import { ResultView } from "./components/ResultView";
import { ErrorView } from "./components/ErrorView";
import { LiveDialog } from "./components/LiveDialog";
import { ReportDialog } from "./components/ReportDialog";
import {
  ApiError,
  cancelJob,
  createJob,
  getHealth,
  getJob,
  inspectSubscription,
  sendPresence,
} from "./api";
import type {
  Aggregate,
  Job,
  JobRecord,
  LiveStats,
  RatioResult,
  SubConfig,
  SubState,
  TestRow,
} from "./types";
import { translate, type Lang, type MessageKey } from "./i18n";
import {
  SELF_HOST_SCRIPT,
  copyText,
  errorKey,
  hostFromUrl,
  isInvalidRatio,
  isNonRoundRatio,
  normalizeUrl,
} from "./utils";
import { shareResult } from "./share";

const ACTIVE_JOB_KEY = "ratio-active-job-react";

class CancelledError extends Error {
  constructor() {
    super("Cancelled");
    this.name = "CancelledError";
  }
}

function sleep(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

function scrollTop() {
  window.scrollTo({
    top: 0,
    behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth",
  });
}

function loadRecord(): JobRecord | null {
  try {
    const raw = sessionStorage.getItem(ACTIVE_JOB_KEY);
    if (!raw) return null;
    const value = JSON.parse(raw) as JobRecord;
    return value && typeof value.id === "string" ? value : null;
  } catch {
    return null;
  }
}

function saveRecord(record: JobRecord) {
  try {
    sessionStorage.setItem(ACTIVE_JOB_KEY, JSON.stringify(record));
  } catch {
    /* quota or private mode */
  }
}

function clearRecord() {
  try {
    sessionStorage.removeItem(ACTIVE_JOB_KEY);
  } catch {
    /* ignore */
  }
}

function clientId() {
  const key = "ratio-client-id";
  try {
    const existing = localStorage.getItem(key);
    if (existing && /^[A-Za-z0-9_-]{8,64}$/.test(existing)) return existing;
    const next =
      typeof crypto.randomUUID === "function"
        ? crypto.randomUUID()
        : `anon-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 12)}`;
    localStorage.setItem(key, next);
    return next;
  } catch {
    return `anon-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 12)}`;
  }
}

function configMeta(config: SubConfig) {
  return {
    index: config.index,
    protocol: config.protocol,
    target: config.target,
    remark: config.remark,
  };
}

function computeAggregate(rows: TestRow[]): Aggregate | null {
  const successes = rows.filter((row) => row.result);
  if (!successes.length) return null;
  const measured = successes.filter((row) => !isInvalidRatio(row.result!.ratio));
  const basis = measured.length ? measured : successes;
  let downloaded = 0;
  let deducted = 0;
  let background = 0;
  let weight = 0;
  let duration = 0;

  basis.forEach((row) => {
    const result = row.result!;
    const bytes = Number(result.downloaded_bytes) || 0;
    const ratio = Number(result.ratio) || 0;
    if (bytes <= 0) return;
    const bg = Number(result.background_bytes_in_test) || 0;
    downloaded += bytes;
    weight += bytes;
    background += bg;
    deducted += isInvalidRatio(ratio)
      ? ratio * bytes
      : Math.max(0, (Number(result.delta_total) || ratio * bytes) - bg);
  });

  successes.forEach((row) => {
    duration += Number(row.result!.duration_seconds) || 0;
  });

  const last = successes[successes.length - 1].result!;
  return {
    ratio: weight > 0 ? deducted / weight : 0,
    downloaded,
    deducted,
    background,
    excluded: successes.length - basis.length,
    count: basis.length,
    duration,
    testedAt: last.tested_at,
    core: last.core,
    xrayVersion: last.xray_version,
    claimed: last.claimed_ratio ?? null,
    nonRound: successes.some((row) =>
      isNonRoundRatio(row.result!.ratio_raw ?? row.result!.ratio),
    ),
  };
}

type View = "landing" | "setup" | "progress" | "result" | "error";

interface BatchState {
  total: number;
  completed: number;
  startedAt: number;
}

interface RunContext {
  url: string;
  host: string;
  configs: Array<Pick<SubConfig, "index" | "protocol" | "target" | "remark">>;
  results: TestRow[];
  completedBase: number;
  total: number;
  startedAt: number;
  size: number;
  bg: boolean;
}

export default function App() {
  const [lang, setLang] = useState<Lang>(() =>
    document.documentElement.lang === "en" ? "en" : "fa",
  );
  const [theme, setTheme] = useState<"light" | "dark">(() =>
    document.documentElement.getAttribute("data-theme") === "dark" ? "dark" : "light",
  );
  const [view, setView] = useState<View>("landing");
  const [url, setUrl] = useState("");
  const [sub, setSub] = useState<SubState | null>(null);
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [size, setSize] = useState(25);
  const [bg, setBg] = useState(false);
  const [inspecting, setInspecting] = useState(false);
  const [inspectError, setInspectError] = useState<{
    message: string;
    detail?: string;
  } | null>(null);
  const [starting, setStarting] = useState(false);
  const [job, setJob] = useState<Job | null>(null);
  const [batch, setBatch] = useState<BatchState | null>(null);
  const [testPicks, setTestPicks] = useState<number[]>([]);
  const [results, setResults] = useState<TestRow[]>([]);
  const [reconnecting, setReconnecting] = useState(false);
  const [canceling, setCanceling] = useState(false);
  const [error, setError] = useState<{
    code: string;
    message: string;
    log: string[];
  } | null>(null);
  const [sharing, setSharing] = useState(false);
  const [brandName, setBrandName] = useState("");
  const [brandImage, setBrandImage] = useState<string | null>(null);
  const [liveOpen, setLiveOpen] = useState(false);
  const [reportOpen, setReportOpen] = useState(false);
  const [live, setLive] = useState<LiveStats | null>(null);
  const [version, setVersion] = useState<string | null>(null);
  const [toast, setToast] = useState<string | null>(null);

  const activeRef = useRef<{ cancelled: boolean; jobId: string | null }>({
    cancelled: false,
    jobId: null,
  });
  const lastJobRef = useRef<Job | null>(null);
  const toastTimer = useRef<number | null>(null);
  const resumedRef = useRef(false);

  const t = useCallback(
    (key: MessageKey, vars?: Record<string, string | number>) =>
      translate(lang, key, vars),
    [lang],
  );

  const showToast = useCallback((message: string) => {
    setToast(message);
    if (toastTimer.current) window.clearTimeout(toastTimer.current);
    toastTimer.current = window.setTimeout(() => setToast(null), 3600);
  }, []);

  useEffect(() => {
    return () => {
      if (toastTimer.current) window.clearTimeout(toastTimer.current);
    };
  }, []);

  useEffect(() => {
    const root = document.documentElement;
    root.lang = lang;
    root.dir = lang === "fa" ? "rtl" : "ltr";
    root.setAttribute("data-theme", theme);
    root.setAttribute("data-mode", theme);
    root.style.colorScheme = theme;
    document.title = t("pageTitle");
    const meta = document.querySelector('meta[name="theme-color"]');
    if (meta) meta.setAttribute("content", theme === "dark" ? "#0b0f16" : "#f8fafc");
    try {
      localStorage.setItem("ratio-lang", lang);
      localStorage.setItem("ratio-theme", theme);
    } catch {
      /* ignore */
    }
  }, [lang, theme, t]);

  useEffect(() => {
    getHealth()
      .then((data) => setVersion(data.version))
      .catch(() => setVersion(null));
  }, []);

  useEffect(() => {
    const id = clientId();
    let alive = true;
    const beat = async () => {
      if (document.visibilityState === "hidden") return;
      try {
        const data = await sendPresence(id);
        if (alive) setLive(data);
      } catch {
        /* best effort */
      }
    };
    void beat();
    const timer = window.setInterval(() => void beat(), 20_000);
    const onVisible = () => {
      if (document.visibilityState === "visible") void beat();
    };
    document.addEventListener("visibilitychange", onVisible);
    return () => {
      alive = false;
      window.clearInterval(timer);
      document.removeEventListener("visibilitychange", onVisible);
    };
  }, []);

  const pollJob = useCallback(async (jobId: string) => {
    let failures = 0;
    while (true) {
      if (activeRef.current.cancelled) throw new CancelledError();
      let current: Job;
      try {
        current = await getJob(jobId);
        failures = 0;
        setReconnecting(false);
      } catch (caught) {
        const code = caught instanceof ApiError ? caught.code : "network";
        const retriable = code === "network" || code === "timeout" || code === "aborted";
        if (retriable && failures < 30) {
          failures += 1;
          setReconnecting(true);
          await sleep(Math.min(12_000, 1_000 * failures));
          continue;
        }
        throw caught;
      }
      lastJobRef.current = current;
      setJob(current);
      if (current.status === "done" && current.result) {
        return { result: current.result, errorCode: undefined, errorMessage: undefined, errorLog: undefined };
      }
      if (current.status === "cancelled") throw new CancelledError();
      if (current.status === "error") {
        return {
          result: null,
          errorCode: current.error_code || "unknown",
          errorMessage: current.error,
          errorLog: current.log,
        };
      }
      await sleep(1000);
    }
  }, []);

  const finish = useCallback(
    (rows: TestRow[]) => {
      setResults(rows);
      setJob(null);
      setBatch(null);
      clearRecord();
      setReconnecting(false);
      const successes = rows.filter((row) => row.result);
      if (successes.length) {
        setView("result");
        scrollTop();
        return;
      }
      const last = rows[rows.length - 1];
      setError({
        code: last?.errorCode || "unknown",
        message: last?.errorMessage || "",
        log: last?.errorLog || [],
      });
      setView("error");
      scrollTop();
    },
    [],
  );

  const runPicks = useCallback(
    async (picks: number[], context: RunContext) => {
      let queue = [...picks];
      let rows = [...context.results];
      let completed = context.completedBase;
      const fallback = (pick: number) => ({
        index: pick,
        protocol: "",
        target: "",
        remark: `#${pick}`,
      });

      while (queue.length) {
        if (activeRef.current.cancelled) return { rows, cancelled: true };
        const pick = queue.shift()!;
        const config = context.configs.find((item) => item.index === pick) || fallback(pick);
        let created: { id: string };
        try {
          created = await createJob({
            sub_url: context.url,
            pick,
            size_mb: context.size,
            sync: context.bg,
          });
        } catch (caught) {
          const apiError = caught instanceof ApiError ? caught : null;
          rows = [
            ...rows,
            {
              pick,
              config,
              result: null,
              errorCode: apiError?.code || "unknown",
              errorMessage: apiError?.message || "",
              errorLog: apiError?.log || [],
            },
          ];
          completed += 1;
          setResults(rows);
          setBatch({ total: context.total, completed, startedAt: context.startedAt });
          continue;
        }

        if (activeRef.current.cancelled) {
          void cancelJob(created.id).catch(() => null);
          return { rows, cancelled: true };
        }

        activeRef.current.jobId = created.id;
        saveRecord({
          id: created.id,
          url: context.url,
          host: context.host,
          pick,
          label: config.remark || config.target || `#${pick}`,
          startedAt: context.startedAt,
          size: context.size,
          bg: context.bg,
          total: context.total,
          completed,
          configs: context.configs,
          queue: [...queue],
          results: rows.map((row) => ({
            pick: row.pick,
            result: row.result,
            errorCode: row.errorCode,
            errorMessage: row.errorMessage,
            errorLog: row.errorLog,
          })),
        });

        try {
          const outcome = await pollJob(created.id);
          if (activeRef.current.cancelled) return { rows, cancelled: true };
          rows = [
            ...rows,
            {
              pick,
              config,
              result: outcome.result,
              errorCode: outcome.errorCode,
              errorMessage: outcome.errorMessage,
              errorLog: outcome.errorLog,
            },
          ];
        } catch (caught) {
          if (caught instanceof CancelledError) return { rows, cancelled: true };
          const apiError = caught instanceof ApiError ? caught : null;
          rows = [
            ...rows,
            {
              pick,
              config,
              result: null,
              errorCode: apiError?.code || "unknown",
              errorMessage: apiError?.message || "",
              errorLog: apiError?.log || [],
            },
          ];
        }

        completed += 1;
        activeRef.current.jobId = null;
        setResults(rows);
        setBatch({ total: context.total, completed, startedAt: context.startedAt });
        clearRecord();
      }
      return { rows, cancelled: false };
    },
    [pollJob],
  );

  useEffect(() => {
    if (resumedRef.current) return;
    const record = loadRecord();
    if (!record) return;
    resumedRef.current = true;
    activeRef.current = { cancelled: false, jobId: record.id };
    const configs = record.configs || [];
    const queue = record.queue || [];
    const savedResults = record.results || [];
    const fallback = (pick: number) => ({
      index: pick,
      protocol: "",
      target: "",
      remark: `#${pick}`,
    });

    void (async () => {
      setSub({
        url: record.url || "",
        host: record.host || "",
        count: configs.length,
        configs: configs as SubConfig[],
        usage: null,
        insecure: (record.url || "").toLowerCase().startsWith("http://"),
      });
      setSize(record.size || 25);
      setBg(Boolean(record.bg));
      setSelected(new Set(configs.map((config) => config.index)));
      setTestPicks(configs.map((config) => config.index).sort((a, b) => a - b));
      setBatch({
        total: record.total || 1,
        completed: record.completed || 0,
        startedAt: record.startedAt || Date.now(),
      });
      setView("progress");

      const prior: TestRow[] = savedResults.map((row) => ({
        pick: row.pick,
        config: configs.find((config) => config.index === row.pick) || fallback(row.pick),
        result: row.result,
        errorCode: row.errorCode,
        errorMessage: row.errorMessage,
        errorLog: row.errorLog,
      }));
      setResults(prior);

      let outcome: {
        result: RatioResult | null;
        errorCode?: string;
        errorMessage?: string;
        errorLog?: string[];
      };
      try {
        const current = await getJob(record.id);
        lastJobRef.current = current;
        setJob(current);
        if (current.status === "done" && current.result) {
          outcome = { result: current.result };
        } else if (current.status === "error") {
          outcome = {
            result: null,
            errorCode: current.error_code || "unknown",
            errorMessage: current.error,
            errorLog: current.log,
          };
        } else if (current.status === "cancelled") {
          clearRecord();
          setView("setup");
          return;
        } else {
          outcome = await pollJob(record.id);
        }
      } catch (caught) {
        if (caught instanceof CancelledError) {
          clearRecord();
          setView("setup");
          return;
        }
        if (caught instanceof ApiError && caught.status === 404) {
          clearRecord();
          setView("landing");
          return;
        }
        const apiError = caught instanceof ApiError ? caught : null;
        outcome = {
          result: null,
          errorCode: apiError?.code || "unknown",
          errorMessage: apiError?.message || "",
          errorLog: apiError?.log || [],
        };
      }

      const currentConfig =
        configs.find((config) => config.index === record.pick) || fallback(record.pick);
      const rows = [
        ...prior,
        {
          pick: record.pick,
          config: currentConfig,
          result: outcome.result,
          errorCode: outcome.errorCode,
          errorMessage: outcome.errorMessage,
          errorLog: outcome.errorLog,
        },
      ];
      const completed = (record.completed || 0) + 1;
      activeRef.current.jobId = null;
      setResults(rows);
      setBatch({
        total: record.total || 1,
        completed,
        startedAt: record.startedAt || Date.now(),
      });

      if (queue.length) {
        const rest = await runPicks(queue, {
          url: record.url,
          host: record.host,
          configs,
          results: rows,
          completedBase: completed,
          total: record.total || 1,
          startedAt: record.startedAt || Date.now(),
          size: record.size || 25,
          bg: Boolean(record.bg),
        });
        if (rest.cancelled) {
          setView("setup");
          return;
        }
        finish(rest.rows);
        return;
      }
      finish(rows);
    })();
  }, [finish, pollJob, runPicks]);

  const aggregate = useMemo(() => computeAggregate(results), [results]);
  const currentBatch =
    batch || { total: 1, completed: 0, startedAt: Date.now() };

  const nextName = useMemo(() => {
    if (!sub || !batch) return null;
    const pick = testPicks[batch.completed + 1];
    if (pick === undefined) return null;
    const config = sub.configs.find((item) => item.index === pick);
    return config ? config.remark || config.target || `#${pick}` : null;
  }, [sub, batch, testPicks]);

  async function inspect() {
    const normalized = normalizeUrl(url.trim().split(/\s+/)[0] || "");
    if (!normalized) {
      setInspectError({ message: t("invalidUrl") });
      return;
    }
    setInspecting(true);
    setInspectError(null);
    try {
      const data = await inspectSubscription(normalized);
      const nextSub: SubState = {
        ...data,
        url: normalized,
        host: hostFromUrl(normalized),
      };
      setSub(nextSub);
      setSelected(new Set(data.configs.slice(0, 1).map((config) => config.index)));
      setResults([]);
      setError(null);
      setUrl(normalized);
      setView("setup");
      scrollTop();
    } catch (caught) {
      const apiError = caught instanceof ApiError ? caught : null;
      setInspectError({
        message: t(errorKey(apiError?.code) as MessageKey),
        detail: apiError
          ? [apiError.message, ...apiError.log].filter(Boolean).join("\n")
          : String(caught),
      });
    } finally {
      setInspecting(false);
    }
  }

  function toggleConfig(index: number) {
    if (starting) return;
    setSelected((current) => {
      const next = new Set(current);
      if (next.has(index)) next.delete(index);
      else next.add(index);
      return next;
    });
  }

  function patchConfig(index: number, patch: Partial<SubConfig>) {
    setSub((current) =>
      current
        ? {
            ...current,
            configs: current.configs.map((config) =>
              config.index === index ? { ...config, ...patch } : config,
            ),
          }
        : current,
    );
  }

  async function startTest() {
    if (!sub || starting) return;
    const picks = [...selected].sort((a, b) => a - b);
    if (!picks.length) {
      showToast(t("noSelection"));
      return;
    }
    activeRef.current = { cancelled: false, jobId: null };
    setStarting(true);
    setError(null);
    setResults([]);
    setJob(null);
    setReconnecting(false);
    setCanceling(false);
    setTestPicks(picks);
    const startedAt = Date.now();
    setBatch({ total: picks.length, completed: 0, startedAt });
    setView("progress");
    scrollTop();
    try {
      const outcome = await runPicks(picks, {
        url: sub.url,
        host: sub.host,
        configs: sub.configs
          .filter((config) => selected.has(config.index))
          .map(configMeta),
        results: [],
        completedBase: 0,
        total: picks.length,
        startedAt,
        size,
        bg,
      });
      if (outcome.cancelled) {
        setView("setup");
        return;
      }
      finish(outcome.rows);
    } catch (caught) {
      const apiError = caught instanceof ApiError ? caught : null;
      setError({
        code: apiError?.code || "unknown",
        message: apiError?.message || "",
        log: apiError?.log || [],
      });
      setView("error");
    } finally {
      setStarting(false);
    }
  }

  async function cancelTest() {
    activeRef.current.cancelled = true;
    setCanceling(true);
    const jobId = activeRef.current.jobId;
    try {
      if (jobId) await cancelJob(jobId);
    } catch {
      showToast(t("cancelFailed"));
    } finally {
      setCanceling(false);
    }
    clearRecord();
    activeRef.current.jobId = null;
    setJob(null);
    setBatch(null);
    setResults([]);
    setReconnecting(false);
    setError(null);
    setView("setup");
    scrollTop();
    showToast(t("cancelled"));
  }

  function resetToLanding() {
    activeRef.current.cancelled = true;
    clearRecord();
    setSub(null);
    setSelected(new Set());
    setResults([]);
    setError(null);
    setJob(null);
    setBatch(null);
    setTestPicks([]);
    setUrl("");
    setInspectError(null);
    setView("landing");
    scrollTop();
    window.setTimeout(() => document.getElementById("sub-url")?.focus(), 50);
  }

  async function onShare() {
    if (!aggregate || sharing) return;
    setSharing(true);
    try {
      const outcome = await shareResult({
        aggregate,
        rows: results,
        lang,
        theme,
        brandName,
        brandImage,
      });
      if (outcome === "downloaded") showToast(t("shareFallback"));
      else if (outcome === "failed") showToast(t("shareFailed"));
    } catch {
      showToast(t("shareFailed"));
    } finally {
      setSharing(false);
    }
  }

  async function onSelfHost() {
    showToast((await copyText(SELF_HOST_SCRIPT)) ? t("selfHostCopied") : t("selfHostFailed"));
  }

  function onLogoFile(file: File) {
    if (!file.type.startsWith("image/") || file.size > 5 * 1024 * 1024) {
      showToast(t("brandImageInvalid"));
      return;
    }
    const reader = new FileReader();
    reader.onload = () => setBrandImage(typeof reader.result === "string" ? reader.result : null);
    reader.onerror = () => showToast(t("brandImageInvalid"));
    reader.readAsDataURL(file);
  }

  const openLive = () => {
    setLiveOpen(true);
    void sendPresence(clientId())
      .then((data) => setLive(data))
      .catch(() => null);
  };

  return (
    <TooltipProvider>
      <div className="z-app">
        <Header
          lang={lang}
          theme={theme}
          online={live?.online ?? null}
          onOpenLive={openLive}
          onToggleTheme={() => setTheme((current) => (current === "dark" ? "light" : "dark"))}
          onToggleLang={() => setLang((current) => (current === "fa" ? "en" : "fa"))}
          showHome={view === "setup" || view === "result" || view === "error"}
          onHome={resetToLanding}
        />

        {view === "landing" && (
          <Landing
            lang={lang}
            url={url}
            onUrlChange={(value) => {
              setUrl(value);
              if (inspectError) setInspectError(null);
            }}
            onInspect={() => void inspect()}
            inspecting={inspecting}
            error={inspectError}
            version={version}
            onSelfHost={() => void onSelfHost()}
            onReport={() => setReportOpen(true)}
          />
        )}

        {view === "setup" && sub && (
          <SetupView
            lang={lang}
            sub={sub}
            selected={selected}
            size={size}
            bg={bg}
            starting={starting}
            onToggle={toggleConfig}
            onRemove={resetToLanding}
            onStart={() => void startTest()}
            onSize={setSize}
            onBg={(value) => {
              setBg(value);
              if (value) showToast(t("bgToast"));
            }}
            onPatchConfig={patchConfig}
            onToast={showToast}
          />
        )}

        {view === "progress" && (
          <ProgressView
            lang={lang}
            job={job}
            batch={currentBatch}
            nextName={nextName}
            reconnecting={reconnecting}
            canceling={canceling}
            onCancel={() => void cancelTest()}
          />
        )}

        {view === "result" && aggregate && (
          <ResultView
            lang={lang}
            aggregate={aggregate}
            rows={results}
            sharing={sharing}
            onShare={() => void onShare()}
            onNewSub={resetToLanding}
          onRetest={() => void startTest()}
            onReport={() => setReportOpen(true)}
            brandName={brandName}
            brandImage={brandImage}
            onBrandName={setBrandName}
            onLogoFile={onLogoFile}
            onClearLogo={() => setBrandImage(null)}
          />
        )}

        {view === "error" && error && (
          <ErrorView
            lang={lang}
            code={error.code}
            message={error.message}
            log={error.log}
            canRetry={Boolean(sub)}
            onRetry={() => void startTest()}
            onChooseConfig={() => {
              setView("setup");
              scrollTop();
            }}
            onNewSub={resetToLanding}
          />
        )}

        {toast && (
          <div className="z-toast" role="status" aria-live="polite">
            {toast}
          </div>
        )}

        <LiveDialog open={liveOpen} onOpenChange={setLiveOpen} lang={lang} live={live} />
        <ReportDialog
          open={reportOpen}
          onOpenChange={setReportOpen}
          lang={lang}
          subUrl={sub?.url || ""}
          result={aggregate}
          onSent={showToast}
        />
      </div>
    </TooltipProvider>
  );
}
