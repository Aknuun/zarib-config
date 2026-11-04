import { Button } from "@cloudflare/kumo";
import { House, Moon, Sun, Translate } from "@phosphor-icons/react";
import type { Lang } from "../i18n";
import { translate } from "../i18n";

export function LogoMark({ size = 32 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 32 32" aria-hidden="true">
      <rect width="32" height="32" rx="9" fill="var(--color-kumo-brand)" />
      <path
        d="M8 22h16M11 22v-4M16 22v-7M21 22v-5"
        stroke="#fff"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  );
}

interface HeaderProps {
  lang: Lang;
  theme: "light" | "dark";
  online: number | null;
  onOpenLive: () => void;
  onToggleTheme: () => void;
  onToggleLang: () => void;
  onHome?: () => void;
  showHome?: boolean;
}

export function Header({
  lang,
  theme,
  online,
  onOpenLive,
  onToggleTheme,
  onToggleLang,
  onHome,
  showHome = false,
}: HeaderProps) {
  const t = (key: Parameters<typeof translate>[1]) => translate(lang, key);
  return (
    <header className="z-header">
      <div className="z-container flex h-14 items-center justify-between gap-2">
        <a href="/" className="z-brand" aria-label={t("appName")}>
          <LogoMark />
          <span className="flex min-w-0 flex-col leading-tight">
            <span className="text-[15px] font-bold">{t("appName")}</span>
            <span className="hidden text-[12px] font-normal text-kumo-subtle sm:block">
              {t("tagline")}
            </span>
          </span>
        </a>
        <div className="flex items-center gap-1">
          <button
            type="button"
            className="z-live-pill"
            onClick={onOpenLive}
            aria-label={`${t("liveStats")}: ${online ?? "—"} ${t("online")}`}
          >
            <span className="z-live-dot" aria-hidden="true" />
            <span className="z-num text-[13px] font-semibold">{online ?? "—"}</span>
            <span className="hidden text-[12px] text-kumo-subtle min-[380px]:inline">
              {t("online")}
            </span>
          </button>
          <Button
            variant="ghost"
            shape="square"
            className="size-8"
            icon={theme === "dark" ? <Sun size={19} /> : <Moon size={19} />}
            aria-label={theme === "dark" ? t("themeLight") : t("themeDark")}
            title={theme === "dark" ? t("themeLight") : t("themeDark")}
            onClick={onToggleTheme}
          />
          <Button
            variant="ghost"
            shape="square"
            className="size-8"
            icon={<Translate size={19} />}
            aria-label={t("switchLang")}
            title={t("switchLang")}
            onClick={onToggleLang}
          />
          {showHome && onHome && (
            <Button
              variant="ghost"
              shape="square"
              className="size-8"
              icon={<House size={19} />}
              aria-label={t("back")}
              title={t("back")}
              onClick={onHome}
            />
          )}
        </div>
      </div>
    </header>
  );
}
