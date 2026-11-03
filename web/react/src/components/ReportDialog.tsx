import { useEffect, useState, type FormEvent } from "react";
import { Button, Dialog, Input, InputArea } from "@cloudflare/kumo";
import { PaperPlaneTilt, X } from "@phosphor-icons/react";
import type { Aggregate } from "../types";
import type { Lang } from "../i18n";
import { translate, type MessageKey } from "../i18n";
import { sendReport, ApiError } from "../api";

interface ReportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  lang: Lang;
  subUrl: string;
  result: Aggregate | null;
  error?: { code: string; message: string; log: string[] } | null;
  view?: string;
  onSent: (message: string) => void;
}

export function ReportDialog({
  open,
  onOpenChange,
  lang,
  subUrl,
  result,
  error,
  view,
  onSent,
}: ReportDialogProps) {
  const t = (key: MessageKey) => translate(lang, key);
  const [description, setDescription] = useState("");
  const [link, setLink] = useState(subUrl);
  const [status, setStatus] = useState("");
  const [sending, setSending] = useState(false);

  useEffect(() => {
    if (open) {
      setLink(subUrl);
      setStatus("");
    }
  }, [open, subUrl]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!description.trim()) {
      setStatus(t("reportRequired"));
      return;
    }
    setSending(true);
    setStatus("");
    try {
      const nav = navigator as Navigator & {
        userAgentData?: { platform?: string; mobile?: boolean };
        deviceMemory?: number;
      };
      const meta = {
        platform: nav.platform || "",
        ua_platform: nav.userAgentData?.platform || "",
        ua_mobile: nav.userAgentData?.mobile ? "بله" : "خیر",
        language: nav.language || "",
        timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || "",
        screen: `${window.screen.width}x${window.screen.height}`,
        viewport: `${window.innerWidth}x${window.innerHeight}`,
        pixel_ratio: String(window.devicePixelRatio || 1),
        touch: nav.maxTouchPoints > 0 ? "بله" : "خیر",
        cores: String(nav.hardwareConcurrency || ""),
        memory: String(nav.deviceMemory || ""),
        page: `${window.location.pathname}${window.location.hash}`,
        view: view || "",
      };
      await sendReport({
        description: description.trim(),
        sub_url: link.trim(),
        meta,
        view: view || "",
        error_code: error?.code || "",
        error_message: error?.message || "",
        error_log: (error?.log || []).slice(-40),
        result: result
          ? {
              ratio: result.ratio,
              downloaded_bytes: result.downloaded,
              delta_total: result.deducted,
              configs_tested: result.count,
              core: result.core,
              xray_version: result.xrayVersion,
              tested_at: result.testedAt,
            }
          : null,
      });
      setDescription("");
      setLink("");
      onOpenChange(false);
      onSent(t("reportSent"));
    } catch (error) {
      const code = error instanceof ApiError ? error.code : "unknown";
      setStatus(
        code === "report_disabled"
          ? t("reportDisabled")
          : code === "invalid_url"
            ? t("reportInvalidLink")
            : code === "rate_limited"
              ? t("reportRateLimited")
              : code === "description_too_long"
                ? t("reportTooLong")
                : code === "invalid_input"
                  ? t("reportRequired")
                  : t("reportFailed"),
      );
    } finally {
      setSending(false);
    }
  }

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog size="lg" className="z-dialog p-5 sm:p-6">
        <div className="flex items-start justify-between gap-3">
          <div>
            <Dialog.Title className="text-[18px] font-semibold">
              {t("reportHeading")}
            </Dialog.Title>
            <Dialog.Description className="mt-1 text-[13px] text-kumo-subtle">
              {t("reportPrivacy")}
            </Dialog.Description>
          <p className="mt-2 text-[12px] text-kumo-subtle">{t("reportMetaNote")}</p>
          {error && <p className="mt-1 text-[12px] text-kumo-subtle">{t("reportErrorAttached")}</p>}
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

        <form className="z-report-form" noValidate onSubmit={submit}>
          <div>
            <label htmlFor="report-description" className="z-field-label">
              {t("reportDescription")}
            </label>
            <InputArea
              id="report-description"
              aria-label={t("reportDescription")}
              rows={4}
              maxLength={1500}
              value={description}
              placeholder={t("reportDescriptionPlaceholder")}
              onChange={(event) => setDescription(event.target.value)}
              className="min-h-28 w-full text-[16px] sm:text-[14px]"
            />
          </div>
          <div>
            <label htmlFor="report-link" className="z-field-label">
              {t("reportSub")}
            </label>
            <Input
              id="report-link"
              aria-label={t("reportSub")}
              dir="ltr"
              type="url"
              maxLength={1500}
              value={link}
              className="h-11 w-full text-[16px] sm:text-[14px]"
              placeholder="https://…"
              onChange={(event) => setLink(event.target.value)}
            />
            <p className="z-hint mt-1">{t("reportSubHint")}</p>
          </div>
          {status && (
            <p className="z-inline-error" role="alert">
              {status}
            </p>
          )}
          <div className="flex flex-wrap gap-2 pt-1">
            <Button
              type="submit"
              variant="primary"
              className="min-h-11 px-4 text-[13px]"
              icon={<PaperPlaneTilt size={17} />}
              loading={sending}
            >
              {sending ? t("reportSending") : t("reportSend")}
            </Button>
            <Dialog.Close
              render={(props) => (
                <Button {...props} variant="ghost" className="z-btn-ghost min-h-11 min-w-11 px-3 text-[13px]">
                  {t("cancel")}
                </Button>
              )}
            />
          </div>
        </form>
      </Dialog>
    </Dialog.Root>
  );
}
