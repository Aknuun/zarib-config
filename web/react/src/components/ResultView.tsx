import { useRef } from "react";
import { Badge, Button, Input } from "@cloudflare/kumo";
import {
  ArrowClockwise,
  Plus,
  ImageSquare,
  ShareNetwork,
  WarningCircle,
  X,
} from "@phosphor-icons/react";
import type { Aggregate, TestRow } from "../types";
import type { Lang } from "../i18n";
import { translate, type MessageKey } from "../i18n";
import {
  errorKey,
  formatBytes,
  formatDate,
  formatDuration,
  formatRatio,
  isInvalidRatio,
  isNoRatio,
  ratioNumber,
  ratioPercent,
  ratioTone,
} from "../utils";
import { Ruler } from "./Ruler";

interface ResultViewProps {
  lang: Lang;
  aggregate: Aggregate;
  rows: TestRow[];
  batch: boolean;
  sharing: boolean;
  onShare: () => void;
  onRetest: () => void;
  onNewSub: () => void;
  onReport: () => void;
  brandName: string;
  brandImage: string | null;
  onBrandName: (value: string) => void;
  onLogoFile: (file: File) => void;
  onClearLogo: () => void;
}

export function ResultView({
  lang,
  aggregate,
  rows,
  batch,
  sharing,
  onShare,
  onRetest,
  onNewSub,
  onReport,
  brandName,
  brandImage,
  onBrandName,
  onLogoFile,
  onClearLogo,
}: ResultViewProps) {
  const t = (key: MessageKey, vars?: Record<string, string | number>) =>
    translate(lang, key, vars);
  const fileRef = useRef<HTMLInputElement>(null);
  const tone = ratioTone(aggregate.ratio);
  const invalid = isInvalidRatio(aggregate.ratio);
  const noRatio = isNoRatio(aggregate.ratio);
  const high = !invalid && aggregate.ratio >= 1.5;
  const unit = Math.round(aggregate.ratio * 100) === 100;
  const verdict = invalid
    ? t("verdictInvalid")
    : noRatio
      ? unit
        ? t("verdictNoRatio")
        : t("verdictClose")
      : t("verdictRatio", { n: ratioNumber(aggregate.ratio) });

  const claimed = aggregate.claimed ?? 0;
  const hasClaim = aggregate.claimed !== null && claimed >= 0.1;
  const claimedPct = claimed > 0 ? ((aggregate.ratio - claimed) / claimed) * 100 : 0;
  const claimedMatches = Math.abs(claimedPct) < 5;

  const notes: string[] = [];
  if (aggregate.background > 0) {
    notes.push(t("backgroundNote", { n: formatBytes(aggregate.background) }));
  }
  if (aggregate.excluded > 0) {
    notes.push(t("excludedNote", { n: aggregate.excluded }));
  }
  if (aggregate.nonRound) notes.push(t("nonRoundNote"));

  return (
    <main className="z-container z-page z-result-page">
      <div className={`z-result-grid${batch ? " is-batch" : ""}`}>
        <section className="z-tile z-result-card" aria-live="polite">
          <h1 className="sr-only">{t("resultEyebrow")}</h1>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <p className="z-eyebrow">{t("resultEyebrow")}</p>
            <div className="flex flex-wrap gap-1.5">
              {batch ? (
                <Badge variant="secondary">{t("batchBadge", { n: aggregate.count })}</Badge>
              ) : aggregate.count > 1 ? (
                <Badge variant="secondary">{t("averageNote", { n: aggregate.count })}</Badge>
              ) : null}
              {invalid ? (
                <Badge variant="error">{t("badgeInvalid")}</Badge>
              ) : noRatio ? (
                <Badge variant="success">{t("badgeNoRatio")}</Badge>
              ) : high ? (
                <Badge variant="warning">{t("badgeHigh")}</Badge>
              ) : null}
            </div>
          </div>

          <div className={`z-coef z-tone-${tone}`} dir="ltr">
            <span>{ratioNumber(aggregate.ratio)}</span>
            <span className="z-coef-unit">×</span>
          </div>

          <Ruler
            percent={ratioPercent(aggregate.ratio)}
            tone={tone}
            labels
            ariaLabel={t("rulerAria")}
            className="z-result-ruler"
          />

          <p className="z-verdict">{verdict}</p>

          {hasClaim && (
            <div
              className={`z-claimed${
                claimedMatches ? " is-match" : claimedPct < 0 ? " is-worse" : " is-better"
              }`}
            >
              <span className="z-claimed-title">
                {t("claimedTitle", { c: ratioNumber(claimed) })}
              </span>
              <span className="z-claimed-diff">
                {claimedMatches
                  ? t("claimedRealMatch", { r: ratioNumber(aggregate.ratio) })
                  : claimedPct < 0
                    ? t("claimedRealLower", {
                        r: ratioNumber(aggregate.ratio),
                        p: Math.abs(claimedPct).toFixed(1),
                      })
                    : t("claimedRealHigher", {
                        r: ratioNumber(aggregate.ratio),
                        p: claimedPct.toFixed(1),
                      })}
              </span>
            </div>
          )}

          {batch && <p className="z-hint">{t("batchNote")}</p>}

          {notes.length > 0 && (
            <ul className="z-notes">
              {notes.map((note) => (
                <li key={note}>
                  <WarningCircle size={14} aria-hidden="true" />
                  <span>{note}</span>
                </li>
              ))}
            </ul>
          )}
        </section>

        {!batch && (
          <section
            className="z-tile z-tile--perconfig z-result-list-card"
            aria-labelledby="per-config-title"
          >
            <div className="flex items-baseline justify-between gap-2">
              <h2 id="per-config-title" className="text-[14px] font-semibold">
                {t("perConfig")}
              </h2>
              <span className="z-num text-[12px] text-kumo-subtle">
                {formatDate(aggregate.testedAt, lang)}
              </span>
            </div>
            <div className="z-result-list">
              {rows.map((row) => {
                const name = row.config.remark || `#${row.pick}`;
                const ratio = row.result ? Number(row.result.ratio) || 0 : null;
                const rowTone = ratio === null ? "danger" : ratioTone(ratio);
                const deducted = row.result
                  ? Math.max(
                      0,
                      (Number(row.result.delta_total) || 0) -
                        (Number(row.result.background_bytes_in_test) || 0),
                    )
                  : 0;
                return (
                  <div key={`${row.pick}-${name}`} className="z-result-row">
                    <div className="flex items-center justify-between gap-3">
                      <span className="min-w-0 truncate text-[13px] font-medium">{name}</span>
                      {ratio === null ? (
                        <span className="shrink-0 text-[12px] text-kumo-danger">
                          {t("invalidRow")}
                        </span>
                      ) : (
                        <span className="flex shrink-0 items-baseline gap-1.5">
                          {isInvalidRatio(ratio) && (
                            <span className="text-[12px] text-kumo-danger">
                              {t("invalidRow")}
                            </span>
                          )}
                          <span className={`z-num z-result-row-ratio z-tone-${rowTone}`}>
                            {formatRatio(ratio)}
                          </span>
                        </span>
                      )}
                    </div>
                    {ratio !== null && (
                      <span className="z-num mt-0.5 block text-[12px] text-kumo-subtle" dir="ltr">
                        {formatBytes(row.result?.downloaded_bytes || 0)} →{" "}
                        {formatBytes(deducted)}
                      </span>
                    )}
                    {row.errorCode && (
                      <span className="mt-0.5 block text-[12px] text-kumo-danger">
                        {t(errorKey(row.errorCode) as MessageKey)}
                      </span>
                    )}
                  </div>
                );
              })}
            </div>
          </section>
        )}
      </div>

      <section className="z-tile z-tile--stats">
        <dl className="z-result-stats">
          <div>
            <dt>{t("downloaded")}</dt>
            <dd className="z-num">{formatBytes(aggregate.downloaded)}</dd>
          </div>
          <div>
            <dt>{t("deducted")}</dt>
            <dd className="z-num">{formatBytes(aggregate.deducted)}</dd>
          </div>
          <div>
            <dt>{t("duration")}</dt>
            <dd className="z-num">{formatDuration(aggregate.duration)}</dd>
          </div>
        </dl>
      </section>

      <section className="z-tile z-tile--actions">
        <div className="z-result-actions">
          <div className="z-action-grid">
            <Button
              variant="secondary"
              size="lg"
              className="col-span-2 h-12 w-full justify-center text-[14px] sm:col-span-1"
              icon={<ArrowClockwise size={18} />}
              onClick={onRetest}
            >
              {t("retest")}
            </Button>
            <Button
              variant="secondary"
              size="lg"
              className="h-12 w-full justify-center text-[14px]"
              icon={<ShareNetwork size={18} />}
              loading={sharing}
              onClick={onShare}
            >
              {sharing ? t("sharing") : t("share")}
            </Button>
            <Button
              variant="secondary"
              size="lg"
              className="h-12 w-full justify-center text-[14px]"
              icon={<Plus size={18} />}
              onClick={onNewSub}
            >
              {t("newSub")}
            </Button>
            <Button
              variant="ghost"
              size="lg"
              className="col-span-2 h-12 w-full justify-center text-[14px] text-kumo-subtle sm:col-span-1"
              icon={<WarningCircle size={18} />}
              onClick={onReport}
            >
              {t("report")}
            </Button>
          </div>
        </div>
      </section>

      <details className="z-tile z-brand-details">
        <summary>
          <span className="flex items-center gap-2">
            <ImageSquare size={18} aria-hidden="true" />
            {t("brandFold")}
          </span>
        </summary>
        <div className="z-brand-body">
          <div className="flex items-start gap-3">
            <Button
              variant="outline"
              shape="square"
              className="size-14 shrink-0 overflow-hidden p-0"
              onClick={() => fileRef.current?.click()}
              aria-label={brandImage ? t("logoAlt") : t("addLogo")}
            >
              {brandImage ? (
                <img src={brandImage} alt={t("logoAlt")} className="size-full object-cover" />
              ) : (
                <ImageSquare size={22} />
              )}
            </Button>
            <input
              ref={fileRef}
              type="file"
              accept="image/*"
              aria-label={t("addLogo")}
              tabIndex={-1}
              className="sr-only"
              onChange={(event) => {
                const file = event.target.files?.[0];
                if (file) onLogoFile(file);
                event.target.value = "";
              }}
            />
            <div className="min-w-0 flex-1">
              <p className="z-hint">{t("brandHint")}</p>
              {brandImage && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="mt-1 min-h-11 px-3 text-[12px] text-kumo-danger"
                  icon={<X size={14} />}
                  onClick={onClearLogo}
                >
                  {t("clearLogo")}
                </Button>
              )}
            </div>
          </div>
          <div>
            <label htmlFor="brand-name" className="z-field-label">
              {t("brandName")}
            </label>
            <Input
              id="brand-name"
              aria-label={t("brandName")}
              size="sm"
              maxLength={60}
              value={brandName}
              className="h-11 w-full text-[16px] sm:text-[14px]"
              placeholder={t("brandNamePlaceholder")}
              onChange={(event) => onBrandName(event.target.value)}
            />
          </div>
        </div>
      </details>

      <p className="z-disclaimer">{t("disclaimer")}</p>
    </main>
  );
}
