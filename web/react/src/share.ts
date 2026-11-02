import type { Aggregate, TestRow } from "./types";
import type { Lang } from "./i18n";
import { translate, type MessageKey } from "./i18n";
import {
  copyText,
  formatBytes,
  formatDate,
  formatRatio,
  isInvalidRatio,
  isNoRatio,
  ratioNumber,
  ratioPercent,
  ratioTone,
  type RatioTone,
} from "./utils";

interface ShareOptions {
  aggregate: Aggregate;
  rows: TestRow[];
  lang: Lang;
  theme: "light" | "dark";
  brandName: string;
  brandImage: string | null;
}

const COLORS = {
  light: {
    bg: "#f8fafc",
    card: "#ffffff",
    text: "#0f172a",
    muted: "#64748b",
    line: "rgba(15, 23, 42, 0.18)",
    brand: "#2563eb",
    good: "#0f766e",
    warn: "#b45309",
    danger: "#b91c1c",
  },
  dark: {
    bg: "#0b0f16",
    card: "#121722",
    text: "#f1f5f9",
    muted: "#94a3b8",
    line: "rgba(241, 245, 249, 0.2)",
    brand: "#3b82f6",
    good: "#2dd4bf",
    warn: "#fbbf24",
    danger: "#f87171",
  },
};

function toneColor(tone: RatioTone, colors: (typeof COLORS)["light"]) {
  return tone === "danger" ? colors.danger : tone === "warn" ? colors.warn : colors.brand;
}

function roundRect(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  width: number,
  height: number,
  radius: number,
) {
  ctx.beginPath();
  if (typeof ctx.roundRect === "function") {
    ctx.roundRect(x, y, width, height, radius);
    return;
  }
  ctx.moveTo(x + radius, y);
  ctx.arcTo(x + width, y, x + width, y + height, radius);
  ctx.arcTo(x + width, y + height, x, y + height, radius);
  ctx.arcTo(x, y + height, x, y, radius);
  ctx.arcTo(x, y, x + width, y, radius);
  ctx.closePath();
}

function font(ctx: CanvasRenderingContext2D, weight: number, size: number) {
  ctx.font = `${weight} ${size}px Estedad, Tahoma, sans-serif`;
}

interface DrawTextOptions {
  size?: number;
  weight?: number;
  color?: string;
  align?: CanvasTextAlign;
  rtl?: boolean;
}

function drawText(
  ctx: CanvasRenderingContext2D,
  text: string,
  x: number,
  y: number,
  options: DrawTextOptions = {},
) {
  const {
    size = 26,
    weight = 400,
    color = "#0f172a",
    align = "left",
    rtl = false,
  } = options;
  font(ctx, weight, size);
  ctx.fillStyle = color;
  ctx.textAlign = align;
  ctx.textBaseline = "alphabetic";
  ctx.direction = rtl ? "rtl" : "ltr";
  ctx.fillText(text, x, y);
}

function wrapText(
  ctx: CanvasRenderingContext2D,
  text: string,
  maxWidth: number,
  size: number,
  weight = 400,
) {
  font(ctx, weight, size);
  const words = text.split(/\s+/).filter(Boolean);
  const lines: string[] = [];
  let line = "";
  for (const word of words) {
    const candidate = line ? `${line} ${word}` : word;
    if (ctx.measureText(candidate).width <= maxWidth) {
      line = candidate;
    } else {
      if (line) lines.push(line);
      line = word;
    }
  }
  if (line) lines.push(line);
  return lines;
}

function drawRuler(
  ctx: CanvasRenderingContext2D,
  x: number,
  width: number,
  y: number,
  percent: number,
  color: string,
  lineColor: string,
  labels: boolean,
  marker = true,
) {
  ctx.strokeStyle = lineColor;
  ctx.lineWidth = 1;
  for (let index = 0; index <= 60; index += 1) {
    const tickX = x + (width / 60) * index;
    const long = index % 10 === 0;
    const medium = index % 5 === 0;
    const height = long ? 22 : medium ? 15 : 9;
    ctx.beginPath();
    ctx.moveTo(tickX, y);
    ctx.lineTo(tickX, y - height);
    ctx.stroke();
  }
  if (marker) {
    const markerX = x + (width * percent) / 100;
    ctx.strokeStyle = color;
    ctx.lineWidth = 3;
    ctx.beginPath();
    ctx.moveTo(markerX, y - 30);
    ctx.lineTo(markerX, y + 2);
    ctx.stroke();
    ctx.fillStyle = color;
    ctx.beginPath();
    ctx.arc(markerX, y - 32, 5, 0, Math.PI * 2);
    ctx.fill();
  }
  if (labels) {
    drawText(ctx, "1×", x + width * 0.1666, y + 32, {
      size: 20,
      color: lineColor,
      align: "center",
    });
    drawText(ctx, "2×", x + width * 0.8333, y + 32, {
      size: 20,
      color: lineColor,
      align: "center",
    });
  }
}

function loadImage(src: string) {
  return new Promise<HTMLImageElement>((resolve, reject) => {
    const image = new Image();
    image.onload = () => resolve(image);
    image.onerror = reject;
    image.src = src;
  });
}

function renderCanvas(
  options: ShareOptions,
  brandImage: HTMLImageElement | null,
): HTMLCanvasElement {
  const { aggregate, rows, lang, theme, brandName } = options;
  const t = (key: MessageKey, vars?: Record<string, string | number>) =>
    translate(lang, key, vars);
  const colors = COLORS[theme];
  const rtl = lang === "fa";
  const tone = ratioTone(aggregate.ratio);
  const accent = toneColor(tone, colors);
  const W = 1080;
  const pad = 72;
  const listRows = rows.slice(0, 6);
  const listStart = 920;
  const listEnd = listStart + listRows.length * 78;
  const H = Math.max(1350, listEnd + 190);

  const canvas = document.createElement("canvas");
  canvas.width = W;
  canvas.height = H;
  const ctx = canvas.getContext("2d");
  if (!ctx) return canvas;

  ctx.fillStyle = colors.bg;
  ctx.fillRect(0, 0, W, H);

  drawText(ctx, t("appName"), rtl ? W - pad : pad, 96, {
    size: 30,
    weight: 700,
    color: colors.text,
    align: rtl ? "right" : "left",
    rtl,
  });
  drawText(ctx, t("shareSite"), rtl ? W - pad : pad, 138, {
    size: 18,
    color: colors.muted,
    align: rtl ? "right" : "left",
  });

  // Signature ruler across the top
  drawRuler(ctx, pad, W - pad * 2, 210, 0, colors.brand, colors.line, false, false);

  drawText(ctx, t("resultEyebrow"), rtl ? W - pad : pad, 300, {
    size: 22,
    weight: 500,
    color: colors.muted,
    align: rtl ? "right" : "left",
    rtl,
  });

  const ratioText = ratioNumber(aggregate.ratio);
  font(ctx, 600, 168);
  const ratioWidth = ctx.measureText(ratioText).width;
  font(ctx, 500, 72);
  const timesWidth = ctx.measureText("×").width;
  const groupWidth = ratioWidth + 6 + timesWidth;
  const groupX = rtl ? W - pad - groupWidth : pad;
  drawText(ctx, ratioText, groupX, 480, {
    size: 168,
    weight: 600,
    color: accent,
    align: "left",
  });
  drawText(ctx, "×", groupX + ratioWidth + 6, 480, {
    size: 72,
    weight: 500,
    color: accent,
    align: "left",
  });

  const logoSize = 200;
  const logoX = rtl ? pad : W - pad - logoSize;
  const logoY = 292;
  ctx.save();
  ctx.shadowColor = "rgba(0, 0, 0, 0.22)";
  ctx.shadowBlur = 32;
  ctx.shadowOffsetY = 12;
  roundRect(ctx, logoX, logoY, logoSize, logoSize, 28);
  ctx.fillStyle = colors.card;
  ctx.fill();
  ctx.restore();
  ctx.strokeStyle = colors.line;
  ctx.lineWidth = 2;
  roundRect(ctx, logoX, logoY, logoSize, logoSize, 28);
  ctx.stroke();
  if (brandImage) {
    ctx.save();
    roundRect(ctx, logoX + 6, logoY + 6, logoSize - 12, logoSize - 12, 22);
    ctx.clip();
    ctx.drawImage(brandImage, logoX + 6, logoY + 6, logoSize - 12, logoSize - 12);
    ctx.restore();
  } else {
    const tile = 24;
    roundRect(ctx, logoX + tile, logoY + tile, logoSize - tile * 2, logoSize - tile * 2, 22);
    ctx.fillStyle = colors.brand;
    ctx.fill();
    ctx.strokeStyle = "#fff";
    ctx.lineWidth = 8;
    ctx.lineCap = "round";
    const bx = logoX + 52;
    const by = logoY + 142;
    ctx.beginPath();
    ctx.moveTo(bx, by);
    ctx.lineTo(bx + 96, by);
    ctx.moveTo(bx + 18, by);
    ctx.lineTo(bx + 18, by - 28);
    ctx.moveTo(bx + 48, by);
    ctx.lineTo(bx + 48, by - 50);
    ctx.moveTo(bx + 78, by);
    ctx.lineTo(bx + 78, by - 36);
    ctx.stroke();
  }
  if (brandName.trim()) {
    drawText(ctx, brandName.trim(), logoX + logoSize / 2, logoY + logoSize + 32, {
      size: 24,
      weight: 600,
      color: colors.text,
      align: "center",
      rtl,
    });
  }

  drawRuler(
    ctx,
    pad,
    W - pad * 2,
    540,
    ratioPercent(aggregate.ratio),
    accent,
    colors.line,
    true,
  );

  const invalid = isInvalidRatio(aggregate.ratio);
  const noRatio = isNoRatio(aggregate.ratio);
  const unit = Math.round(aggregate.ratio * 100) === 100;
  const verdict = invalid
    ? t("verdictInvalid")
    : noRatio
      ? unit
        ? t("verdictNoRatio")
        : t("verdictClose")
      : t("verdictRatio", { n: ratioNumber(aggregate.ratio) });
  const verdictLines = wrapText(ctx, verdict, W - pad * 2, 30, 500);
  verdictLines.slice(0, 2).forEach((line, index) => {
    drawText(ctx, line, rtl ? W - pad : pad, 630 + index * 44, {
      size: 30,
      weight: 500,
      color: colors.text,
      align: rtl ? "right" : "left",
      rtl,
    });
  });

  const statsY = 710;
  const gap = 20;
  const boxW = (W - pad * 2 - gap) / 2;
  const stats = [
    { label: t("shareDownloadedLabel"), value: formatBytes(aggregate.downloaded) },
    { label: t("shareDeductedLabel"), value: formatBytes(aggregate.deducted) },
  ];
  stats.forEach((stat, index) => {
    const x = rtl ? W - pad - boxW - index * (boxW + gap) : pad + index * (boxW + gap);
    roundRect(ctx, x, statsY, boxW, 120, 20);
    ctx.fillStyle = colors.card;
    ctx.fill();
    ctx.strokeStyle = colors.line;
    ctx.lineWidth = 1.5;
    ctx.stroke();
    drawText(ctx, stat.label, rtl ? x + boxW - 24 : x + 24, statsY + 44, {
      size: 20,
      color: colors.muted,
      align: rtl ? "right" : "left",
      rtl,
    });
    drawText(ctx, stat.value, rtl ? x + boxW - 24 : x + 24, statsY + 88, {
      size: 30,
      weight: 600,
      color: colors.text,
      align: rtl ? "right" : "left",
    });
  });

  if (listRows.length > 0) {
    drawText(ctx, t("perConfig"), rtl ? W - pad : pad, listStart - 28, {
      size: 22,
      weight: 600,
      color: colors.muted,
      align: rtl ? "right" : "left",
      rtl,
    });
    listRows.forEach((row, index) => {
      const y = listStart + index * 78;
      const name = row.config.remark || row.config.target || `#${row.pick}`;
      const ratio = row.result ? Number(row.result.ratio) || 0 : null;
      const rowTone = ratio === null ? "danger" : ratioTone(ratio);
      const rowColor = toneColor(rowTone, colors);
      const deducted = row.result
        ? Math.max(
            0,
            (Number(row.result.delta_total) || 0) -
              (Number(row.result.background_bytes_in_test) || 0),
          )
        : 0;
      ctx.strokeStyle = colors.line;
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.moveTo(pad, y + 58);
      ctx.lineTo(W - pad, y + 58);
      ctx.stroke();
      drawText(ctx, name.slice(0, 42), rtl ? W - pad : pad, y + 34, {
        size: 24,
        weight: 500,
        color: colors.text,
        align: rtl ? "right" : "left",
        rtl,
      });
      if (ratio === null) {
        drawText(ctx, t("invalidRow"), rtl ? pad : W - pad, y + 34, {
          size: 22,
          color: colors.danger,
          align: rtl ? "left" : "right",
          rtl,
        });
      } else {
        drawText(ctx, formatRatio(ratio), rtl ? pad : W - pad, y + 34, {
          size: 26,
          weight: 600,
          color: rowColor,
          align: rtl ? "left" : "right",
        });
        drawText(
          ctx,
          `${formatBytes(row.result?.downloaded_bytes || 0)} → ${formatBytes(deducted)}`,
          W / 2,
          y + 34,
          {
            size: 20,
            color: colors.muted,
            align: "center",
          },
        );
      }
    });
  }

  // Footer
  ctx.strokeStyle = colors.line;
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.moveTo(pad, H - 120);
  ctx.lineTo(W - pad, H - 120);
  ctx.stroke();
  drawText(ctx, t("shareFooter"), rtl ? W - pad : pad, H - 72, {
    size: 22,
    color: colors.muted,
    align: rtl ? "right" : "left",
    rtl,
  });
  drawText(ctx, t("shareSite"), rtl ? pad : W - pad, H - 72, {
    size: 22,
    weight: 600,
    color: colors.text,
    align: rtl ? "left" : "right",
  });
  drawText(ctx, formatDate(aggregate.testedAt, lang), rtl ? W - pad : pad, H - 40, {
    size: 17,
    color: colors.muted,
    align: rtl ? "right" : "left",
  });

  return canvas;
}

export async function shareResult(options: ShareOptions): Promise<"shared" | "downloaded" | "failed" | "cancelled"> {
  try {
    await document.fonts.ready;
  } catch {
    /* ignore */
  }
  let brandImage: HTMLImageElement | null = null;
  if (options.brandImage) {
    try {
      brandImage = await loadImage(options.brandImage);
    } catch {
      brandImage = null;
    }
  }
  const canvas = renderCanvas(options, brandImage);
  const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, "image/png"));
  if (!blob) return "failed";
  const t = (key: MessageKey) => translate(options.lang, key);
  const caption = `${t("shareTitle")} — ${formatRatio(options.aggregate.ratio)}\n${t("shareFooter")}\n${t("shareSite")}`;
  const file = new File([blob], "zarib-result.png", { type: "image/png" });

  if (navigator.canShare && navigator.canShare({ files: [file] })) {
    try {
      await navigator.share({ files: [file], text: caption });
      return "shared";
    } catch (error) {
      if (error instanceof DOMException && error.name === "AbortError") return "cancelled";
    }
  }

  try {
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = "zarib-result.png";
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    window.setTimeout(() => URL.revokeObjectURL(url), 2000);
    await copyText(caption);
    return "downloaded";
  } catch {
    return "failed";
  }
}
