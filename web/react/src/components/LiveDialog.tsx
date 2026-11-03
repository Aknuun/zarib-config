import { useState, type ReactNode } from "react";
import { Button, Dialog, Tabs, Tooltip } from "@cloudflare/kumo";
import { Info, X } from "@phosphor-icons/react";
import type { LiveStats } from "../types";
import type { Lang } from "../i18n";
import { translate, type MessageKey } from "../i18n";
import { formatBytes } from "../utils";

interface LiveDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  lang: Lang;
  live: LiveStats | null;
}

type Range = "h24" | "d7" | "d30" | "all";

export function LiveDialog({ open, onOpenChange, lang, live }: LiveDialogProps) {
  const t = (key: MessageKey, vars?: Record<string, string | number>) =>
    translate(lang, key, vars);
  const [range, setRange] = useState<Range>("h24");

  const bucket = live
    ? range === "h24"
      ? live.h24
      : range === "d7"
        ? live.d7
        : range === "d30"
          ? live.d30
          : live.overall
    : null;
  const failed = bucket?.failed || {};
  const failedTotal = Number(failed.total) || 0;
  const tests = Number(bucket?.tests) || 0;
  const success = Math.max(0, tests - failedTotal);
  const successRate = tests > 0 ? `${success} (${Math.round((success / tests) * 100)}%)` : "0";

  const row = (label: ReactNode, value: string) => (
    <div className="z-live-row" key={typeof label === "string" ? label : value}>
      <dt>{label}</dt>
      <dd className="z-num" dir="ltr">{value}</dd>
    </div>
  );

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog size="lg" className="z-dialog p-4 sm:p-6">
        <div className="flex items-start justify-between gap-3">
          <div>
            <Dialog.Title className="text-[18px] font-semibold">
              {t("liveHeading")}
            </Dialog.Title>
            <Dialog.Description className="mt-1 text-[13px] text-kumo-subtle">
              {t("liveTitle")}
            </Dialog.Description>
          </div>
          <Dialog.Close
            render={(props) => (
              <Button
                {...props}
                variant="ghost"
                shape="square"
                className="z-btn-ghost size-11 shrink-0"
                icon={<X size={19} />}
                aria-label={t("close")}
              />
            )}
          />
        </div>

        <Tabs
          value={range}
          onValueChange={(value) => setRange(value as Range)}
          tabs={[
            { value: "h24", label: t("last24h") },
            { value: "d7", label: t("last7d") },
            { value: "d30", label: t("last30d") },
            { value: "all", label: t("all") },
          ]}
          variant="segmented"
          size="sm"
          listClassName="h-11! [&_button]:my-0 [&_button]:min-h-11 [&_button]:min-w-11"
          className="z-live-tabs mt-3"
        />

        <div className="z-dialog-tiles mt-3">
          <section className="z-dialog-tile">
            <h3 className="z-live-group">{t("traffic")}</h3>
            <dl>
              {row(t("statDownloaded"), live ? formatBytes(bucket?.downloaded || 0) : "—")}
              {row(t("statDeducted"), live ? formatBytes(bucket?.deducted || 0) : "—")}
            </dl>
          </section>

          <section className="z-dialog-tile">
            <h3 className="z-live-group">{t("tests")}</h3>
            <dl>
              {row(t("statTests"), live ? String(tests) : "—")}
              {row(t("statSuccess"), live ? successRate : "—")}
              {row(t("statFailed"), live ? String(failedTotal) : "—")}
              {row(t("statWithRatio"), live ? String(bucket?.with_ratio || 0) : "—")}
              {row(t("statNormal"), live ? String(bucket?.normal || 0) : "—")}
              {row(t("statInvalid"), live ? String(bucket?.invalid || 0) : "—")}
            </dl>
          </section>

          <section className="z-dialog-tile">
            <h3 className="z-live-group">{t("quality")}</h3>
            <dl>
              {row(t("statOnline"), live ? String(live.online) : "—")}
          {row(t("statUnique"), live ? (bucket?.unique_ips != null ? String(bucket.unique_ips) : "—") : "—")}
              {row(
                <Tooltip
                  content={t("proxyHint")}
                  render={
                    <span tabIndex={0} className="inline-flex items-center gap-1">
                      {t("statProxy")}
                      <Info size={13} aria-hidden="true" />
                    </span>
                  }
                />,
                live ? String(bucket?.proxy || 0) : "—",
              )}
            </dl>
            {live && (
              <div className="z-live-lines">
                <p>
                  {t("tunnelLine", {
                    t: bucket?.tunnel || 0,
                    n: bucket?.non_tunnel || 0,
                    u: bucket?.unknown_tunnel || 0,
                  })}
                </p>
                <p>
                  {t("failLine", {
                    c: failed.connect || 0,
                    f: failed.fetch || 0,
                    x: failed.core || 0,
                    t: failed.timeout || 0,
                    u: failed.usage || 0,
                    o: failed.other || 0,
                  })}
                </p>
              </div>
            )}
          </section>
        </div>

        <p className="z-hint mt-3">{t("liveHint")}</p>
      </Dialog>
    </Dialog.Root>
  );
}
