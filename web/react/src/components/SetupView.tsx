import { useState } from "react";
import { Badge, Button, DropdownMenu, Slider, Switch } from "@cloudflare/kumo";
import {
  CheckCircle,
  Circle,
  Copy,
  DotsThree,
  GlobeHemisphereWest,
  Play,
  Pulse,
  Trash,
  WarningCircle,
  XCircle,
} from "@phosphor-icons/react";
import type { SubConfig, SubState } from "../types";
import type { Lang } from "../i18n";
import { translate } from "../i18n";
import { geoTarget, pingTarget } from "../api";
import { copyText } from "../utils";

interface SetupViewProps {
  lang: Lang;
  sub: SubState;
  selected: Set<number>;
  size: number;
  bg: boolean;
  starting: boolean;
  onToggle: (index: number) => void;
  onRemove: () => void;
  onStart: () => void;
  onStartAll: () => void;
  onSize: (value: number) => void;
  onBg: (value: boolean) => void;
  onPatchConfig: (index: number, patch: Partial<SubConfig>) => void;
  onToast: (message: string) => void;
}

export function SetupView({
  lang,
  sub,
  selected,
  size,
  bg,
  starting,
  onToggle,
  onRemove,
  onStart,
  onStartAll,
  onSize,
  onBg,
  onPatchConfig,
  onToast,
}: SetupViewProps) {
  const t = (key: Parameters<typeof translate>[1], vars?: Record<string, string | number>) =>
    translate(lang, key, vars);
  const [busy, setBusy] = useState<Record<number, "ping" | "tunnel">>({});
  const [bulk, setBulk] = useState<"ping" | "tunnel" | null>(null);


  async function runPing(config: SubConfig) {
    if (busy[config.index]) return;
    setBusy((current) => ({ ...current, [config.index]: "ping" }));
    try {
      const data = await pingTarget(config.target);
      onPatchConfig(config.index, { pingMs: data.ok ? data.latency_ms : null });
      if (!data.ok) onToast(t("noResponse"));
    } catch {
      onToast(t("pingFailed"));
    } finally {
      setBusy((current) => {
        const next = { ...current };
        delete next[config.index];
        return next;
      });
    }
  }

  async function runTunnel(config: SubConfig) {
    if (busy[config.index]) return;
    setBusy((current) => ({ ...current, [config.index]: "tunnel" }));
    try {
      const data = await geoTarget(config.target);
      onPatchConfig(config.index, { tunnel: data.tunnel, geoCountry: data.country });
      onToast(data.tunnel ? t("tunnelDetected") : t("tunnelNotDetected"));
    } catch {
      onToast(t("tunnelFailed"));
    } finally {
      setBusy((current) => {
        const next = { ...current };
        delete next[config.index];
        return next;
      });
    }
  }

  async function copyAll() {
    const text = sub.configs
      .map((config) => config.uri)
      .filter(Boolean)
      .join("\n\n");
    if (!text) {
      onToast(t("copyAllFailed"));
      return;
    }
    onToast((await copyText(text)) ? t("copyAllDone") : t("copyAllFailed"));
  }

  async function runPingAll() {
    if (bulk) return;
    setBulk("ping");
    let ok = 0;
    await Promise.allSettled(
      sub.configs.map(async (config) => {
        try {
          const data = await pingTarget(config.target);
          onPatchConfig(config.index, { pingMs: data.ok ? data.latency_ms : null });
          if (data.ok) ok += 1;
        } catch {
          onPatchConfig(config.index, { pingMs: null });
        }
      }),
    );
    setBulk(null);
    onToast(t("pingAllDone", { ok, n: sub.configs.length }));
  }

  async function runTunnelAll() {
    if (bulk) return;
    setBulk("tunnel");
    let ok = 0;
    await Promise.allSettled(
      sub.configs.map(async (config) => {
        try {
          const data = await geoTarget(config.target);
          onPatchConfig(config.index, { tunnel: data.tunnel, geoCountry: data.country });
          if (data.tunnel) ok += 1;
        } catch {
          /* ignore one host */
        }
      }),
    );
    setBulk(null);
    onToast(t("tunnelAllDone", { ok, n: sub.configs.length }));
  }

  return (
    <main className="z-container z-page z-setup-page">
      <div className="z-page-head">
        <div>
          <p className="z-eyebrow">{t("setupEyebrow")}</p>
          <h1 className="z-page-title">{t("setupTitle")}</h1>
          <p className="z-page-sub">{t("setupSub")}</p>
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="min-h-11 min-w-11 shrink-0 px-2 text-[12px] text-kumo-danger"
          icon={<Trash size={16} />}
          aria-label={t("removeSub")}
          onClick={onRemove}
          disabled={starting}
        >
          <span className="max-sm:hidden">{t("removeSub")}</span>
        </Button>
      </div>

      <section className="z-card z-setup-card">
        <div className="z-sub-summary">
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <strong className="z-num truncate text-[15px]" dir="ltr">
                {sub.host}
              </strong>
              <Badge variant="secondary">{t("configCount", { n: sub.count })}</Badge>
            </div>
          </div>
        </div>

        {sub.insecure && (
          <p className="z-warning-line">
            <WarningCircle size={16} aria-hidden="true" />
            {t("insecure")}
          </p>
        )}

        <div className="z-setup-body">
          <div className="z-setup-list">
        <div className="z-list-head">
          <div className="flex items-baseline gap-2">
            <h2 className="text-[14px] font-semibold">{t("configs")}</h2>
            <span className="text-[12px] text-kumo-subtle">
              {t("selectedN", { n: selected.size })}
            </span>
          </div>
        </div>

        <div className="z-bulk-actions">
          <Button
            variant="secondary"
            size="sm"
            className="min-h-11"
            icon={<Copy size={16} />}
            onClick={copyAll}
          >
            {t("copyAll")}
          </Button>
          <Button
            variant="secondary"
            size="sm"
            className="min-h-11"
            icon={<Pulse size={16} />}
            loading={bulk === "ping"}
            disabled={bulk !== null || starting}
            onClick={runPingAll}
          >
            {t("pingAll")}
          </Button>
          <Button
            variant="secondary"
            size="sm"
            className="min-h-11"
            icon={<GlobeHemisphereWest size={16} />}
            loading={bulk === "tunnel"}
            disabled={bulk !== null || starting}
            onClick={runTunnelAll}
          >
            {t("tunnelAll")}
          </Button>
        </div>

        <div className="z-config-list" role="group" aria-label={t("configs")}>
          {sub.configs.map((config) => {
            const isSelected = selected.has(config.index);
            const isBusy = busy[config.index];
            const name = config.remark || config.target || `#${config.index}`;
            return (
              <div
                key={config.index}
                className={`z-config-row${isSelected ? " is-selected" : ""}`}
              >
                <button
                  type="button"
                  role="checkbox"
                  aria-checked={isSelected}
                  className="z-config-main"
                  onClick={() => onToggle(config.index)}
                  disabled={starting}
                >
                  <span className="z-config-check" aria-hidden="true">
                    {isSelected ? (
                      <CheckCircle size={21} weight="fill" />
                    ) : (
                      <Circle size={21} />
                    )}
                  </span>
                  <span className="min-w-0 flex-1 text-start">
                    <span className="block truncate text-[14px] font-medium" dir="auto">{name}</span>
                  </span>
                  {config.pingMs !== undefined && (
                    <span
                      className={`z-ping-chip shrink-0${
                        config.pingMs === null ? " is-bad" : ""
                      }`}
                      title={
                        config.pingMs === null
                          ? t("noResponse")
                          : t("latencyAria", { n: config.pingMs })
                      }
                    >
                      {config.pingMs === null ? (
                        <XCircle size={14} aria-hidden="true" />
                      ) : (
                        <Pulse size={14} aria-hidden="true" />
                      )}
                      <span className="z-num">
                        {config.pingMs === null ? "—" : `${config.pingMs} ms`}
                      </span>
                    </span>
                  )}
                  {config.tunnel !== undefined && config.tunnel !== null && (
                    <Badge
                      variant={config.tunnel ? "success" : "neutral"}
                    >
                      {config.tunnel ? t("tunnel") : t("normal")}
                    </Badge>
                  )}
                  <span className="sr-only">
                    {isSelected ? t("selected") : t("notSelected")}
                  </span>
                </button>
                <DropdownMenu>
                  <DropdownMenu.Trigger
                    render={
                      <Button
                        variant="ghost"
                        shape="square"
                        className="size-11 shrink-0"
                        icon={<DotsThree size={20} />}
                        aria-label={t("configMenu")}
                        disabled={starting}
                      />
                    }
                  />
                  <DropdownMenu.Content>
                    <DropdownMenu.Item
                      className="min-h-11"
                      icon={<Copy size={17} />}
                      onClick={async () => {
                        if (!config.uri) {
                          onToast(t("copyFailed"));
                          return;
                        }
                        onToast(
                          (await copyText(config.uri))
                            ? t("copyDone")
                            : t("copyFailed"),
                        );
                      }}
                    >
                      {t("copyUri")}
                    </DropdownMenu.Item>
                    <DropdownMenu.Item
                      className="min-h-11"
                      icon={<Pulse size={17} />}
                      disabled={Boolean(isBusy)}
                      onClick={() => runPing(config)}
                    >
                      {t("checkPing")}
                    </DropdownMenu.Item>
                    <DropdownMenu.Item
                      className="min-h-11"
                      icon={<GlobeHemisphereWest size={17} />}
                      disabled={Boolean(isBusy)}
                      onClick={() => runTunnel(config)}
                    >
                      {t("checkTunnel")}
                    </DropdownMenu.Item>
                  </DropdownMenu.Content>
                </DropdownMenu>
              </div>
            );
          })}
        </div>

          </div>

          <div className="z-setup-side">
        <div className="z-settings">
          <div className="z-setting">
            <div className="z-setting-head">
              <span>{t("size")}</span>
              <strong className="z-num text-[13px]">{t("sizeValue", { n: size })}</strong>
            </div>
            <div dir="ltr">
              <Slider
                value={size}
                min={5}
                max={100}
                step={5}
                getAriaLabel={() => t("sizeSliderAria")}
                onValueChange={(value) => onSize(Number(value))}
                disabled={starting}
                className="mt-2 w-full"
              />
            </div>
            <p className="z-hint">{t("sizeHint")}</p>
          </div>
          <div className="z-setting">
            <Switch
              checked={bg}
              onCheckedChange={onBg}
              label={t("bg")}
              size="sm"
              disabled={starting}
              aria-label={t("bgSwitchAria")}
            />
            <p className="z-hint">{t("bgHint")}</p>
          </div>
        </div>

        <div className="z-start">
          {selected.size > 1 ? (
            <div className="grid w-full grid-cols-2 gap-2">
              <Button
                variant="primary"
                size="lg"
                className="h-12 w-full justify-center text-[14px]"
                icon={!starting ? <Play size={18} weight="fill" /> : undefined}
                loading={starting}
                disabled={starting}
                onClick={onStart}
              >
                {t("testSerial")}
              </Button>
              <Button
                variant="secondary"
                size="lg"
                className="h-12 w-full justify-center text-[14px]"
                disabled={starting}
                onClick={onStartAll}
              >
                {t("testParallel")}
              </Button>
            </div>
          ) : (
            <Button
              variant="primary"
              size="lg"
              className="h-12 justify-center self-center px-6 text-[14px]"
              icon={!starting ? <Play size={18} weight="fill" /> : undefined}
              loading={starting}
              disabled={selected.size === 0 || starting}
              onClick={onStart}
            >
              {selected.size > 0
                ? t("startN", { n: selected.size })
                : t("start")}
            </Button>
          )}
          <p className="z-start-note">
            <WarningCircle size={14} aria-hidden="true" />
            {t("idleHint")}
          </p>
        </div>
          </div>
        </div>
      </section>
    </main>
  );
}
