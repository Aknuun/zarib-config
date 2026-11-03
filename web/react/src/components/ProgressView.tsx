import { useEffect, useRef, useState } from "react";
import { Button } from "@cloudflare/kumo";
import {
  BookOpen,
  CheckCircle,
  DownloadSimple,
} from "@phosphor-icons/react";
import type { Job } from "../types";
import type { Lang } from "../i18n";
import { localizeLog, translate, type MessageKey } from "../i18n";
import { formatDuration } from "../utils";
import { Ruler } from "./Ruler";

interface ProgressViewProps {
  lang: Lang;
  job: Job | null;
  batch: { total: number; completed: number; startedAt: number };
  batchMode: boolean;
  nextName: string | null;
  reconnecting: boolean;
  canceling: boolean;
  onCancel: () => void;
}

function stepIndex(stage: string, status: string) {
  if (status === "done") return 3;
  if (["queued", "reading", "connecting", "baseline", "drift"].includes(stage)) return 0;
  if (stage === "downloading") return 1;
  if (["waiting", "usage_pending", "waiting_extended"].includes(stage)) return 2;
  return 0;
}

export function ProgressView({
  lang,
  job,
  batch,
  batchMode,
  nextName,
  reconnecting,
  canceling,
  onCancel,
}: ProgressViewProps) {
  const t = (key: MessageKey, vars?: Record<string, string | number>) =>
    translate(lang, key, vars);
  const [now, setNow] = useState(Date.now());
  const [displayPercent, setDisplayPercent] = useState(0);
  const displayRef = useRef(0);
  const targetRef = useRef(0);

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, []);

  const stage = job?.stage || "queued";
  const progress = job?.progress || 0;
  const percent = batchMode
    ? Math.min(99, Math.max(0, progress))
    : Math.min(
        99,
        Math.max(0, ((batch.completed + progress / 100) / Math.max(1, batch.total)) * 100),
      );
  targetRef.current = percent;

  useEffect(() => {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      displayRef.current = percent;
      setDisplayPercent(percent);
      return;
    }
    let raf = 0;
    const tick = () => {
      const target = targetRef.current;
      const next = displayRef.current + (target - displayRef.current) * 0.12;
      if (Math.abs(target - next) < 0.15) {
        displayRef.current = target;
        setDisplayPercent(target);
        return;
      }
      displayRef.current = next;
      setDisplayPercent(next);
      raf = window.requestAnimationFrame(tick);
    };
    raf = window.requestAnimationFrame(tick);
    return () => window.cancelAnimationFrame(raf);
  }, [percent]);

  const activeStep = stepIndex(stage, job?.status || "queued");
  const stageKey = `stage_${stage}` as MessageKey;
  const stageLabel = t(stageKey);
  const rawMessage = job?.message || "";
  const localizedMessage = rawMessage ? localizeLog(lang, rawMessage) : "";
  const message = reconnecting
    ? t("reconnect")
    : localizedMessage && (lang !== "fa" || localizedMessage !== rawMessage)
      ? localizedMessage
      : stageLabel;
  const elapsed = (now - batch.startedAt) / 1000;
  const current = Math.min(batch.total, batch.completed + 1);

  const steps: Array<{ title: MessageKey; icon: React.ReactNode }> = [
    { title: "stepRead", icon: <BookOpen size={18} /> },
    { title: "stepDownload", icon: <DownloadSimple size={18} /> },
    { title: "stepVerify", icon: <CheckCircle size={18} /> },
  ];

  return (
    <main className="z-container z-page z-progress-page">
      <section className="z-tile z-progress-card" aria-labelledby="progress-title">
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="z-eyebrow">{t("progressEyebrow")}</p>
            <h1 id="progress-title" className="z-progress-title">
              {batchMode
                ? t("progressBatch", { n: batch.total })
                : batch.total > 1
                  ? t("progressOf", { i: current, n: batch.total })
                  : t("progressSingle")}
            </h1>
          </div>
          <div className="text-end">
            <div className="z-num z-progress-pct" aria-hidden="true">
              {Math.floor(displayPercent)}%
            </div>
            <div className="z-num text-[12px] text-kumo-subtle">
              {t("elapsed")} {formatDuration(elapsed)}
            </div>
          </div>
        </div>

        <div
          className="z-progress-ruler"
          role="progressbar"
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuenow={Math.floor(displayPercent)}
          aria-label={t("progressEyebrow")}
        >
          <Ruler percent={displayPercent} tone="brand" />
        </div>

        <ol className="z-timeline" aria-label={t("progressEyebrow")}>
          {steps.map((step, index) => {
            const state =
              index < activeStep || activeStep === 3
                ? "done"
                : index === activeStep
                  ? "active"
                  : "pending";
            return (
              <li key={step.title} className={`z-timeline-step is-${state}`}>
                <span className="z-timeline-icon" aria-hidden="true">
                  {state === "done" ? <CheckCircle size={18} weight="fill" /> : step.icon}
                </span>
                <span>{t(step.title)}</span>
                {index < steps.length - 1 && (
                  <span className="z-timeline-line" aria-hidden="true" />
                )}
              </li>
            );
          })}
        </ol>
      </section>

      <section className="z-tile z-stage-card">
        <div className="z-stage">
          <div className="z-stage-head">
            <span className="z-stage-label">{stageLabel}</span>
          </div>
          <p className="sr-only" aria-live="polite" aria-atomic="true">
            {stageLabel} {message !== stageLabel ? message : ""}
          </p>
          {message !== stageLabel && (
            <p className="z-stage-message">{message}</p>
          )}
          {stage === "usage_pending" && (
            <p className="z-hint">{t("usagePendingHint")}</p>
          )}
          {nextName && batch.total > 1 && (
            <p className="z-hint">{t("nextConfig", { name: nextName })}</p>
          )}
        </div>

        <div className="z-progress-actions">
          <Button
            variant="secondary-destructive"
            className="min-h-11 px-4 text-[13px]"
            loading={canceling}
            onClick={onCancel}
          >
            {t("cancelTest")}
          </Button>
        </div>

        <details className="z-details z-log-details">
          <summary>{t("details")}</summary>
          <pre dir="ltr" aria-label={t("logLabel")}>
            {(job?.log || []).map((line) => localizeLog(lang, line)).join("\n") || "—"}
          </pre>
        </details>
      </section>
    </main>
  );
}
