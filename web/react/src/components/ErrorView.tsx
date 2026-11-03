import { Button } from "@cloudflare/kumo";
import { ArrowClockwise, ArrowLeft, PaperPlaneTilt, WarningOctagon } from "@phosphor-icons/react";
import type { Lang } from "../i18n";
import { localizeLog, translate, type MessageKey } from "../i18n";
import { errorKey, suggestionKey } from "../utils";

interface ErrorViewProps {
  lang: Lang;
  code: string;
  message: string;
  log: string[];
  canRetry: boolean;
  onRetry: () => void;
  onChooseConfig: () => void;
  onNewSub: () => void;
  onReport: () => void;
}

export function ErrorView({
  lang,
  code,
  message,
  log,
  canRetry,
  onRetry,
  onChooseConfig,
  onNewSub,
  onReport,
}: ErrorViewProps) {
  const t = (key: MessageKey) => translate(lang, key);
  const detail = [localizeLog(lang, message), ...log.slice(-20).map((line) => localizeLog(lang, line))]
    .filter(Boolean)
    .join("\n")
    .replace(/https?:\/\/\S+/g, "<link>");

  return (
    <main className="z-container z-page z-error-page">
      <section className="z-tile z-error-card" role="alert">
        <span className="z-error-icon" aria-hidden="true">
          <WarningOctagon size={26} />
        </span>
        <h1 className="z-page-title">{t("errorTitle")}</h1>
        <p className="z-error-text">{t(errorKey(code) as MessageKey)}</p>
        <p className="z-page-sub">{t(suggestionKey(code) as MessageKey)}</p>
        <details className="z-details">
          <summary>{t("errorDetails")}</summary>
          <pre dir="ltr">{detail || "—"}</pre>
        </details>
        <div className="z-error-actions">
          {canRetry && (
            <Button
              variant="primary"
              size="lg"
              className="h-12 justify-center text-[14px] max-sm:w-full"
              icon={<ArrowClockwise size={18} />}
              onClick={onRetry}
            >
              {t("retry")}
            </Button>
          )}
          {canRetry && (
            <Button
              variant="secondary"
              size="lg"
              className="h-12 justify-center text-[14px] max-sm:w-full"
              icon={<ArrowLeft size={18} />}
              onClick={onChooseConfig}
            >
              {t("chooseConfig")}
            </Button>
          )}
          <Button
            variant="ghost"
            size="lg"
            className="h-12 justify-center text-[14px] max-sm:w-full"
            icon={<PaperPlaneTilt size={18} />}
            onClick={onReport}
          >
            {t("report")}
          </Button>
          <Button
            variant="ghost"
            size="lg"
            className="h-12 justify-center text-[14px] max-sm:w-full"
            onClick={onNewSub}
          >
            {t("newSub")}
          </Button>
        </div>
      </section>
    </main>
  );
}
