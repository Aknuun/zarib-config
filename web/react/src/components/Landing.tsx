import { Ruler } from "./Ruler";
import { Button, Input } from "@cloudflare/kumo";
import { ArrowLeft, LinkSimple, LockSimple } from "@phosphor-icons/react";
import type { Lang } from "../i18n";
import { translate } from "../i18n";
import { ratioPercent } from "../utils";

interface LandingProps {
  lang: Lang;
  url: string;
  onUrlChange: (value: string) => void;
  onInspect: () => void;
  inspecting: boolean;
  error: { message: string; detail?: string } | null;
  version: string | null;
  onSelfHost: () => void;
  onReport: () => void;
}

export function Landing({
  lang,
  url,
  onUrlChange,
  onInspect,
  inspecting,
  error,
  version,
  onSelfHost,
  onReport,
}: LandingProps) {
  const t = (key: Parameters<typeof translate>[1], vars?: Record<string, string | number>) =>
    translate(lang, key, vars);

  return (
    <div className="z-landing">
      <main className="z-container z-landing-main">
        <section className="z-hero">
          <h1 className="z-hero-title">{t("heroTitle")}</h1>
          <p className="z-hero-sub">{t("heroSub")}</p>
          <div className="z-sample-ruler">
            <div className="flex items-center justify-between text-[12px] text-kumo-subtle">
              <span>{t("sample")}</span>
              <span className="z-num font-semibold text-kumo-default" dir="ltr">
                1.38×
              </span>
            </div>
            <Ruler
              percent={ratioPercent(1.38)}
              tone="warn"
              labels
              ariaLabel={t("rulerAria")}
            />
          </div>
        </section>

        <section className="z-card z-input-card" aria-label={t("subLabel")}>
          <form
            className="flex flex-col gap-3 sm:flex-row sm:items-start"
            noValidate
            onSubmit={(event) => {
              event.preventDefault();
              onInspect();
            }}
          >
            <div className="min-w-0 flex-1">
              <label htmlFor="sub-url" className="z-field-label">
                {t("subLabel")}
              </label>
              <div className="relative">
                <LinkSimple
                  size={18}
                  className="pointer-events-none absolute inset-y-0 left-3 my-auto text-kumo-subtle"
                  aria-hidden="true"
                />
                <Input
                  id="sub-url"
                  size="lg"
                  dir="ltr"
                  className="h-12 w-full pl-10 text-[16px] sm:text-[14px]"
                  type="url"
                  inputMode="url"
                  autoComplete="off"
                  spellCheck={false}
                  placeholder={t("subPlaceholder")}
                  value={url}
                  onChange={(event) => onUrlChange(event.target.value)}
                  aria-label={t("subLabel")}
                  variant={error ? "error" : "default"}
                  aria-invalid={error ? true : undefined}
                  aria-describedby={error ? "sub-error" : undefined}
                />
              </div>
            </div>
            <Button
              type="submit"
              variant="primary"
              size="lg"
              className="h-12 shrink-0 px-5 text-[14px] max-sm:w-full"
              loading={inspecting}
              icon={!inspecting ? <ArrowLeft size={18} /> : undefined}
            >
              {inspecting ? t("inspecting") : t("inspect")}
            </Button>
          </form>

          {error && (
            <div id="sub-error" className="z-inline-error" role="alert">
              <p>{error.message}</p>
              {error.detail && (
                <details className="z-details z-details-tight">
                  <summary>{t("details")}</summary>
                  <pre dir="ltr">{error.detail}</pre>
                </details>
              )}
            </div>
          )}

          <p className="z-privacy">
            <LockSimple size={14} aria-hidden="true" />
            {t("privacy")}
          </p>
        </section>

        <section className="z-how" aria-labelledby="how-title">
          <h2 id="how-title" className="z-eyebrow">
            {t("howTitle")}
          </h2>
          <ol className="z-how-list">
            {(
              [
                ["how1Title", "how1Body"],
                ["how2Title", "how2Body"],
                ["how3Title", "how3Body"],
              ] as const
            ).map(([title, body], index) => (
              <li key={title} className="z-how-item">
                <span className="z-how-num z-num">{index + 1}</span>
                <div>
                  <h3>{t(title)}</h3>
                  <p className="max-sm:hidden">{t(body)}</p>
                </div>
              </li>
            ))}
          </ol>
        </section>
      </main>

      <footer className="z-footer">
        <div className="z-container flex flex-wrap items-center justify-between gap-x-4 gap-y-1">
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              className="min-h-11 px-2 text-[12px]"
              onClick={onSelfHost}
            >
              {t("selfHost")}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="min-h-11 px-2 text-[12px]"
              onClick={onReport}
            >
              {t("report")}
            </Button>
          </div>
          <span className="z-num text-[12px] text-kumo-subtle">
            {version ? `v${version}` : ""}
          </span>
        </div>
      </footer>
    </div>
  );
}
