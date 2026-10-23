(() => {
  "use strict";

  const I18N = {
    fa: {
      pageTitle: "ضریب واقعی ترافیک V2Ray | RatioTest",
      eyebrow: "شفافیت حجم، بدون حدس",
      heroTitle1: "ضریب واقعی",
      heroTitle2: "ترافیک V2Ray",
      heroSub: "۱۰ مگابایت تست، بدون ذخیره لینک.",
      chip1: "مگابایت تست",
      chip2: "بدون ذخیره لینک",
      chip3: "مقایسه با ادعای پنل",
      subTitle: "لینک سابسکریپشن",
      statusReady: "آماده",
      subLabel: "لینک سابسکریپشن",
      loadConfigs: "دریافت کانفیگ‌ها",
      subHint: "لینک فقط برای همین تست استفاده می‌شود و ذخیره نمی‌شود.",
      selectTitle: "انتخاب کانفیگ",
      configLabel: "کانفیگ مورد تست",
      configCount: "{n} کانفیگ",
      configCountTruncated: "{shown} از {total} کانفیگ",
      advanced: "تنظیمات بیشتر",
      sizeLabel: "حجم تست",
      claimedLabel: "ضریب ادعاشده (اختیاری)",
      coreLabel: "هسته Xray",
      coreAuto: "خودکار",
      coreLatest: "جدید",
      coreLegacy: "قدیمی",
      coreSystem: "سیستم",
      coreHint: "اگر کانفیگ با هسته جدید وصل نشد، هسته قدیمی را امتحان کن.",
      coreUsed: "هسته Xray",
      chooseAnother: "کانفیگ دیگر",
      claimOptional: "اختیاری",
      claimDetected: "ادعای این کانفیگ: {n}",
      agreeText: "می‌دانم حدود {mb} مگابایت مصرف می‌شود.",
      startTest: "شروع تست",
      privacyHint: "لینک و نتیجه فقط در حافظه نگهداری می‌شوند و بعد از پایان حذف می‌شوند.",
      progressTitle: "در حال تست…",
      cancel: "لغو",
      showLog: "جزئیات",
      hideLog: "بستن",
      stage_queued: "در صف…",
      stage_reading: "خواندن اطلاعات سابسکریپشن…",
      stage_connecting: "اتصال به کانفیگ…",
      stage_baseline: "خواندن مصرف پایه…",
      stage_drift: "اندازه‌گیری مصرف پس‌زمینه…",
      stage_downloading: "دانلود داده تست…",
      stage_waiting: "انتظار برای ثبت مصرف در پنل…",
      stage_done: "تمام شد",
      stage_cancelled: "لغو شد",
      stage_error: "خطا",
      resultEyebrow: "نتیجه تست",
      sentence: "برای هر ۱ مگابایت دانلود، {n} مگابایت از سابسکریپشن شما کم می‌شود.",
      downloaded: "دانلود واقعی",
      deducted: "کسرشده",
      claimed: "ادعای پنل",
      remaining: "باقی‌مانده",
      notClaimed: "نامشخص",
      verdictMatch: "با ادعا مطابقت دارد",
      verdictHigher: "بیشتر از ادعا",
      verdictLower: "کمتر از ادعا",
      realEquivalent: "معادل واقعی {v}",
      background: "اصلاح پس‌زمینه",
      raw: "خام",
      duration: "مدت تست",
      configUsed: "کانفیگ",
      testAnother: "کانفیگ دیگر",
      shareResult: "اشتراک‌گذاری",
      sharePreparing: "در حال ساخت تصویر…",
      shareFallback: "تصویر نتیجه دانلود شد و متن اشتراک‌گذاری کپی شد. آن را در تلگرام بفرست.",
      shareDownloaded: "تصویر نتیجه دانلود شد. متن اشتراک‌گذاری را دستی کپی کن.",
      shareFailed: "ساخت یا اشتراک‌گذاری تصویر ممکن نشد.",
      shareCaptionTitle: "نتیجه تست ضریب V2Ray من",
      shareFooter: "تست ضریب واقعی V2Ray",
      themeToDark: "تغییر به تم تاریک",
      themeToLight: "تغییر به تم روشن",
      errorTitle: "تست انجام نشد",
      details: "جزئیات",
      tryAgain: "تلاش دوباره",
      howTitle: "چطور کار می‌کند؟",
      how1: "مصرف سابسکریپشن قبل از تست خوانده می‌شود.",
      how2: "حجم مشخصی از داده از طریق کانفیگ دانلود می‌شود.",
      how3: "مصرف بعد از تست خوانده و با حجم واقعی مقایسه می‌شود.",
      howNote: "اگر پنل شما مصرف را با تاخیر ثبت کند، ممکن است نتیجه دقیق نباشد. لینک شما هرگز ذخیره نمی‌شود.",
      footer: "ساخته‌شده برای شفافیت حجم.",
      disclaimer: "فقط برای تست سابسکریپشن خودت.",
      unnamed: "(بدون نام)",
      noUsage: "این سابسکریپشن آمار مصرف ندارد؛ ممکن است تست انجام نشود.",
      usageRemaining: "حجم باقی‌مانده",
      usageUsed: "مصرف‌شده",
      usageTotal: "کل حجم",
      usageOfTotal: "از {v}",
      daysLeft: "{n} روز باقی‌مانده",
      expired: "منقضی شده",
      suggestTryAnother: "کانفیگ دیگری را انتخاب کن، هسته Xray را عوض کن، یا چند دقیقه بعد دوباره تلاش کن.",
      suggestCheckLink: "لینک سابسکریپشن را بررسی کن و مطمئن شو از اینترنت در دسترس است.",
      suggestWait: "کمی صبر کن و دوباره تلاش کن.",
      suggestPanelDelay: "پنل ممکن است مصرف را با تاخیر ثبت کند. یک بار دیگر تست کن یا کانفیگ دیگری را امتحان کن.",
      suggestTryAgain: "دوباره تلاش کن؛ اگر مشکل ادامه داشت کانفیگ دیگری را امتحان کن.",
      err_invalid_url: "لینک معتبر نیست. باید با http:// یا https:// شروع شود.",
      err_blocked_url: "این لینک به آدرس عمومی اشاره نمی‌کند و مجاز نیست.",
      err_fetch_failed: "خواندن سابسکریپشن ممکن نشد. لینک را بررسی کن یا بعداً دوباره تلاش کن.",
      err_no_configs: "کانفیگ پشتیبانی‌شده‌ای در این ساب پیدا نشد.",
      err_probe_failed: "کانفیگ وصل نشد. ممکن است نود خاموش، مسدود یا خیلی کند باشد.",
      err_resolve_failed: "آدرس سرور کانفیگ پیدا نشد (DNS). کانفیگ دیگری را امتحان کن.",
      err_xray_failed: "Xray نتوانست کانفیگ را اجرا کند. جزئیات را ببین.",
      err_usage_unchanged: "پنل مصرف را ثبت نکرد. ممکن است تاخیر داشته باشد یا آمار مصرف ارائه نکند.",
      err_timeout: "تست بیش از حد طول کشید و متوقف شد. نود کند است یا حجم تست زیاد است.",
      err_rate_limited: "تعداد درخواست‌ها زیاد است. کمی بعد دوباره تلاش کن.",
      err_busy: "سرور در حال حاضر شلوغ است. چند دقیقه دیگر تلاش کن.",
      err_cancelled: "تست لغو شد.",
      err_pick_out_of_range: "شماره کانفیگ معتبر نیست؛ دوباره کانفیگ‌ها را دریافت کن.",
      err_invalid_input: "ورودی معتبر نیست.",
      err_not_found: "تست پیدا نشد؛ ممکن است سرور ری‌استارت شده باشد.",
      err_unknown: "خطای پیش‌بینی‌نشده‌ای رخ داد.",
    },
    en: {
      pageTitle: "V2Ray Ratio Test | RatioTest",
      eyebrow: "Transparent data accounting",
      heroTitle1: "Real V2Ray",
      heroTitle2: "traffic ratio",
      heroSub: "A 10 MB test. Link never stored.",
      chip1: "MB test",
      chip2: "Link never stored",
      chip3: "Compares with the claim",
      subTitle: "Subscription link",
      statusReady: "Ready",
      subLabel: "Subscription URL",
      loadConfigs: "Load configs",
      subHint: "The link is used only for this test and is never stored.",
      selectTitle: "Choose a config",
      configLabel: "Config to test",
      configCount: "{n} configs",
      configCountTruncated: "{shown} of {total} configs",
      advanced: "More settings",
      sizeLabel: "Test size",
      claimedLabel: "Claimed ratio (optional)",
      coreLabel: "Xray core",
      coreAuto: "Auto",
      coreLatest: "Latest",
      coreLegacy: "Legacy",
      coreSystem: "System",
      coreHint: "If a config does not connect with the latest core, try the legacy core.",
      coreUsed: "Xray core",
      chooseAnother: "Another config",
      claimOptional: "Optional",
      claimDetected: "This config claims: {n}",
      agreeText: "I know about {mb} MB will be used.",
      startTest: "Start test",
      privacyHint: "The link and result are kept in memory only and are removed after the test.",
      progressTitle: "Testing…",
      cancel: "Cancel",
      showLog: "Details",
      hideLog: "Close",
      stage_queued: "Queued…",
      stage_reading: "Reading subscription…",
      stage_connecting: "Connecting through the config…",
      stage_baseline: "Reading baseline usage…",
      stage_drift: "Measuring background usage…",
      stage_downloading: "Downloading test data…",
      stage_waiting: "Waiting for the panel to record usage…",
      stage_done: "Done",
      stage_cancelled: "Cancelled",
      stage_error: "Error",
      resultEyebrow: "Test result",
      sentence: "For every 1 MB downloaded, {n} MB is deducted from your subscription.",
      downloaded: "Downloaded",
      deducted: "Deducted",
      claimed: "Claimed",
      remaining: "Remaining",
      notClaimed: "Not found",
      verdictMatch: "Matches the claim",
      verdictHigher: "Higher than the claim",
      verdictLower: "Lower than the claim",
      realEquivalent: "real equivalent {v}",
      background: "Background correction",
      raw: "raw",
      duration: "Duration",
      configUsed: "Config",
      testAnother: "Another config",
      shareResult: "Share",
      sharePreparing: "Creating image…",
      shareFallback: "The result image was downloaded and the share text was copied. Attach it in Telegram.",
      shareDownloaded: "The result image was downloaded. Copy the share text manually.",
      shareFailed: "Could not create or share the result image.",
      shareCaptionTitle: "My V2Ray ratio test result",
      shareFooter: "Real V2Ray traffic ratio test",
      themeToDark: "Switch to dark theme",
      themeToLight: "Switch to light theme",
      errorTitle: "Test failed",
      details: "Details",
      tryAgain: "Try again",
      howTitle: "How it works",
      how1: "Subscription usage is read before the test.",
      how2: "A known amount of data is downloaded through the config.",
      how3: "Usage is read again and compared with the actual download.",
      howNote: "If your panel records usage with a delay, the result may be inaccurate. Your link is never stored.",
      footer: "Built for transparent data accounting.",
      disclaimer: "Only test your own subscription.",
      unnamed: "(unnamed)",
      noUsage: "This subscription does not expose usage; the test may not work.",
      usageRemaining: "Remaining",
      usageUsed: "Used",
      usageTotal: "Total",
      usageOfTotal: "of {v}",
      daysLeft: "{n} days left",
      expired: "Expired",
      suggestTryAnother: "Try another config, choose a different Xray core, or retry in a few minutes.",
      suggestCheckLink: "Check the subscription link and make sure it is reachable.",
      suggestWait: "Wait a moment and try again.",
      suggestPanelDelay: "The panel may record usage with a delay. Retry or try another config.",
      suggestTryAgain: "Try again. If the problem continues, try another config.",
      err_invalid_url: "The link is not valid. It must start with http:// or https://.",
      err_blocked_url: "This link does not point to a public address and is not allowed.",
      err_fetch_failed: "Could not read the subscription. Check the link or try again later.",
      err_no_configs: "No supported configs were found in this subscription.",
      err_probe_failed: "The config did not connect. The node may be down, blocked, or too slow.",
      err_resolve_failed: "The config server hostname could not be resolved (DNS). Try another config.",
      err_xray_failed: "Xray could not start the config. See the technical details.",
      err_usage_unchanged: "The panel did not record usage. It may be delayed or may not expose usage.",
      err_timeout: "The test took too long and was stopped. The node is slow or the size is too large.",
      err_rate_limited: "Too many requests. Try again later.",
      err_busy: "The server is busy. Try again in a few minutes.",
      err_cancelled: "The test was cancelled.",
      err_pick_out_of_range: "The config number is invalid; load the configs again.",
      err_invalid_input: "Invalid input.",
      err_not_found: "The test was not found; the server may have restarted.",
      err_unknown: "An unexpected error occurred.",
    },
  };

  const $ = (sel, root = document) => root.querySelector(sel);
  const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));
  const SITE_URL = "https://zarib-config.videobazi.com";
  const THEME_KEY = "ratio-theme";

  const state = {
    lang: document.documentElement.lang === "en" ? "en" : "fa",
    subUrl: "",
    configs: [],
    usage: null,
    configCount: 0,
    selected: null,
    size: 10,
    claimedTouched: false,
    jobId: null,
    pollTimer: null,
    startedAt: 0,
    lastJobData: null,
    logOpen: false,
    result: null,
    error: null,
    errorCode: null,
    errorLog: [],
  };

  try {
    const saved = localStorage.getItem("ratio-lang");
    if (saved === "fa" || saved === "en") state.lang = saved;
  } catch (_) {
    /* ignore */
  }

  function t(key, vars) {
    let value = (I18N[state.lang] && I18N[state.lang][key]) ?? I18N.fa[key] ?? key;
    if (vars) {
      for (const [name, val] of Object.entries(vars)) {
        value = value.replaceAll(`{${name}}`, String(val));
      }
    }
    return value;
  }

  function show(el) {
    if (el) el.classList.remove("hidden");
  }

  function hide(el) {
    if (el) el.classList.add("hidden");
  }

  function formatBytes(value) {
    if (value === null || value === undefined || Number.isNaN(Number(value))) return "—";
    let n = Number(value);
    const units = ["B", "KB", "MB", "GB", "TB", "PB"];
    let index = 0;
    while (Math.abs(n) >= 1024 && index < units.length - 1) {
      n /= 1024;
      index += 1;
    }
    return `${n.toFixed(index === 0 ? 0 : 2)} ${units[index]}`;
  }

  function formatRatio(value) {
    if (value === null || value === undefined || Number.isNaN(Number(value))) return "—";
    return `${Number(value).toFixed(2)}x`;
  }

  function formatDuration(seconds) {
    const total = Math.max(0, Math.round(Number(seconds) || 0));
    const mins = Math.floor(total / 60);
    const secs = total % 60;
    return `${String(mins).padStart(2, "0")}:${String(secs).padStart(2, "0")}`;
  }

  function formatDate(iso) {
    if (!iso) return "—";
    const date = new Date(iso);
    if (Number.isNaN(date.getTime())) return iso;
    try {
      return date.toLocaleString(state.lang === "fa" ? "fa-IR" : "en-GB", {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
    } catch (_) {
      return date.toISOString();
    }
  }

  function toast(message, duration = 2400) {
    const el = document.createElement("div");
    el.className = "toast";
    el.textContent = message;
    document.body.appendChild(el);
    window.setTimeout(() => el.remove(), duration);
  }

  function errorMessage(code) {
    const key = `err_${code || "unknown"}`;
    return I18N[state.lang] && I18N[state.lang][key] ? t(key) : t("err_unknown");
  }

  function errorSuggestion(code) {
    switch (code) {
      case "probe_failed":
      case "resolve_failed":
      case "timeout":
      case "xray_failed":
        return t("suggestTryAnother");
      case "fetch_failed":
      case "blocked_url":
      case "invalid_url":
      case "no_configs":
        return t("suggestCheckLink");
      case "usage_unchanged":
        return t("suggestPanelDelay");
      case "rate_limited":
      case "busy":
        return t("suggestWait");
      default:
        return t("suggestTryAgain");
    }
  }

  function applyTheme(theme, persist) {
    const next = theme === "dark" ? "dark" : "light";
    document.documentElement.setAttribute("data-theme", next);
    const meta = document.querySelector('meta[name="theme-color"]');
    if (meta) meta.setAttribute("content", next === "dark" ? "#050810" : "#f5f8fc");
    const button = $("#themeBtn");
    if (button) {
      const label = t(next === "dark" ? "themeToLight" : "themeToDark");
      button.setAttribute("aria-label", label);
      button.setAttribute("title", label);
    }
    if (persist) {
      try {
        localStorage.setItem(THEME_KEY, next);
      } catch (_) {
        /* ignore */
      }
    }
  }

  function toggleTheme() {
    const current =
      document.documentElement.getAttribute("data-theme") === "dark" ? "dark" : "light";
    applyTheme(current === "dark" ? "light" : "dark", true);
  }

  function applyI18n() {
    document.documentElement.lang = state.lang;
    document.documentElement.dir = state.lang === "fa" ? "rtl" : "ltr";
    document.title = t("pageTitle");
    $("#langBtn").textContent = state.lang === "fa" ? "EN" : "فارسی";
    applyTheme(document.documentElement.getAttribute("data-theme"), false);
    $$("[data-i18n]").forEach((el) => {
      el.textContent = t(el.dataset.i18n);
    });
    updateAgreeText();
    updateLogToggle();
  }

  function setLoading(button, loading) {
    if (!button) return;
    const spinner = $(".spinner", button);
    if (loading) {
      button.disabled = true;
      if (spinner) show(spinner);
    } else {
      if (spinner) hide(spinner);
    }
  }

  function clearInputError() {
    const box = $("#inputError");
    box.textContent = "";
    box.title = "";
    hide(box);
  }

  function showInputError(message, technical) {
    const box = $("#inputError");
    box.textContent = message;
    box.title = technical || "";
    show(box);
  }

  async function api(path, options = {}) {
    const response = await fetch(path, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...(options.headers || {}),
      },
    });
    let data = null;
    try {
      data = await response.json();
    } catch (_) {
      data = null;
    }
    if (!response.ok) {
      const error = new Error((data && data.error) || response.statusText || "Request failed");
      error.code = (data && data.code) || "unknown";
      error.status = response.status;
      error.log = (data && data.log) || [];
      throw error;
    }
    return data;
  }

  function normalizeUrl(raw) {
    let value = raw.trim();
    if (!value) return "";
    if (!/^[a-z][a-z0-9+.-]*:\/\//i.test(value)) value = `https://${value}`;
    return value;
  }

  async function inspect() {
    clearInputError();
    const normalized = normalizeUrl($("#subUrl").value);
    if (!normalized) {
      showInputError(errorMessage("invalid_url"));
      return;
    }
    $("#subUrl").value = normalized;
    state.subUrl = normalized;
    const button = $("#inspectBtn");
    setLoading(button, true);
    hide($("#selectCard"));
    hide($("#resultCard"));
    hide($("#errorCard"));
    hide($("#progressCard"));
    try {
      const data = await api("/api/inspect", {
        method: "POST",
        body: JSON.stringify({ sub_url: normalized }),
      });
      state.configs = data.configs || [];
      state.configCount = data.count || state.configs.length;
      state.usage = data.usage || null;
      state.selected = state.configs.length ? state.configs[0].index : null;
      state.claimedTouched = false;
      $("#claimed").value = "";
      renderUsage();
      renderConfigs();
      renderClaim();
      updateStartState();
      show($("#selectCard"));
      $("#selectCard").scrollIntoView({ behavior: "smooth", block: "start" });
    } catch (error) {
      showInputError(errorMessage(error.code), error.message);
    } finally {
      setLoading(button, false);
      button.disabled = false;
    }
  }

  function renderUsage() {
    const box = $("#usageBox");
    box.replaceChildren();
    const usage = state.usage;
    if (!usage) {
      const top = document.createElement("div");
      top.className = "usage-top";
      const label = document.createElement("span");
      label.textContent = t("usageRemaining");
      const value = document.createElement("strong");
      value.textContent = "—";
      top.append(label, value);
      const hint = document.createElement("p");
      hint.className = "hint small";
      hint.textContent = t("noUsage");
      box.append(top, hint);
      return;
    }

    const top = document.createElement("div");
    top.className = "usage-top";
    const label = document.createElement("span");
    label.textContent = t("usageRemaining");
    const value = document.createElement("strong");
    value.textContent = usage.remaining === null ? "—" : formatBytes(usage.remaining);
    const total = document.createElement("span");
    total.textContent =
      usage.total === null ? "" : t("usageOfTotal", { v: formatBytes(usage.total) });
    top.append(label, value, total);

    const bar = document.createElement("div");
    bar.className = "usage-bar";
    const fill = document.createElement("span");
    const percent = usage.total ? Math.min(100, (usage.used / usage.total) * 100) : 0;
    fill.style.width = `${percent.toFixed(1)}%`;
    bar.append(fill);

    const meta = document.createElement("div");
    meta.className = "usage-meta";
    const used = document.createElement("span");
    used.textContent = `${t("usageUsed")}: ${formatBytes(usage.used)}`;
    meta.append(used);
    if (usage.total !== null) {
      const totalEl = document.createElement("span");
      totalEl.textContent = `${t("usageTotal")}: ${formatBytes(usage.total)}`;
      meta.append(totalEl);
    }
    if (usage.days_left !== null && usage.days_left !== undefined) {
      const days = document.createElement("span");
      days.textContent =
        usage.days_left > 0 ? t("daysLeft", { n: usage.days_left }) : t("expired");
      meta.append(days);
    }
    box.append(top, bar, meta);
  }

  function renderConfigs() {
    const list = $("#configList");
    list.replaceChildren();
    state.configs.forEach((config) => {
      const item = document.createElement("label");
      item.className = "config-item";
      if (config.index === state.selected) item.classList.add("selected");

      const radio = document.createElement("input");
      radio.type = "radio";
      radio.name = "config";
      radio.className = "config-radio";
      radio.value = String(config.index);
      radio.checked = config.index === state.selected;
      radio.addEventListener("change", () => selectConfig(config.index));

      const body = document.createElement("div");
      body.className = "config-body";

      const top = document.createElement("div");
      top.className = "config-top";
      const proto = document.createElement("span");
      proto.className = "proto";
      proto.textContent = config.protocol || "json";
      const target = document.createElement("span");
      target.className = "config-target";
      target.textContent = config.target || "";
      top.append(proto, target);
      if (config.claimed !== null && config.claimed !== undefined) {
        const claim = document.createElement("span");
        claim.className = "config-claim";
        claim.textContent = formatRatio(config.claimed);
        top.append(claim);
      }

      const remark = document.createElement("div");
      remark.className = "config-remark";
      remark.textContent = config.remark || t("unnamed");

      body.append(top, remark);
      item.append(radio, body);
      list.append(item);
    });

    const count = $("#configCount");
    const total = state.configCount || state.configs.length;
    count.textContent =
      total > state.configs.length
        ? t("configCountTruncated", { shown: state.configs.length, total })
        : t("configCount", { n: total });
  }

  function selectConfig(index) {
    state.selected = index;
    $$("#configList .config-item").forEach((item) => {
      const radio = $(".config-radio", item);
      const active = Number(radio.value) === index;
      item.classList.toggle("selected", active);
      radio.checked = active;
    });
    renderClaim();
    updateStartState();
  }

  function selectedConfig() {
    return state.configs.find((config) => config.index === state.selected) || null;
  }

  function renderClaim() {
    const config = selectedConfig();
    const hint = $("#claimHint");
    if (config && config.claimed !== null && config.claimed !== undefined) {
      hint.textContent = t("claimDetected", { n: formatRatio(config.claimed) });
      if (!state.claimedTouched) $("#claimed").value = String(config.claimed);
    } else {
      hint.textContent = t("claimOptional");
      if (!state.claimedTouched) $("#claimed").value = "";
    }
  }

  function updateAgreeText() {
    $("#agreeText").textContent = t("agreeText", { mb: state.size });
  }

  function updateLogToggle() {
    $("#logToggle").textContent = state.logOpen ? t("hideLog") : t("showLog");
  }

  function updateStartState() {
    const agree = $("#agree").checked;
    const hasConfig = Boolean(state.selected);
    $("#startBtn").disabled = !(agree && hasConfig) || Boolean(state.jobId);
  }

  function setSize(size) {
    state.size = Number(size);
    $$("#sizeGroup .seg").forEach((button) => {
      button.classList.toggle("is-active", Number(button.dataset.size) === state.size);
    });
    updateAgreeText();
  }

  async function startTest() {
    if ($("#startBtn").disabled) return;
    clearInputError();
    const claimedRaw = $("#claimed").value.trim();
    let claimed = null;
    if (claimedRaw) {
      claimed = Number(claimedRaw);
      if (Number.isNaN(claimed) || claimed < 0.1 || claimed > 100) {
        showInputError(errorMessage("invalid_input"));
        return;
      }
    }
    const button = $("#startBtn");
    setLoading(button, true);
    button.disabled = true;
    hide($("#errorCard"));
    hide($("#resultCard"));
    try {
      const data = await api("/api/jobs", {
        method: "POST",
        body: JSON.stringify({
          sub_url: state.subUrl,
          pick: state.selected,
          size_mb: state.size,
          claimed,
          wait: 20,
          drift: 10,
          core: $("#coreSelect").value || "auto",
        }),
      });
      state.jobId = data.id;
      state.startedAt = Date.now();
      state.lastJobData = null;
      state.logOpen = false;
      state.result = null;
      state.error = null;
      hide($("#selectCard"));
      resetProgress();
      show($("#progressCard"));
      $("#progressCard").scrollIntoView({ behavior: "smooth", block: "start" });
      startPolling();
    } catch (error) {
      showError(error);
    } finally {
      setLoading(button, false);
      updateStartState();
    }
  }

  function resetProgress() {
    $("#stageLabel").textContent = t("stage_queued");
    $("#stageMessage").textContent = "";
    $("#progressPct").textContent = "0%";
    $("#progressBar").style.width = "0%";
    $("#elapsed").textContent = "00:00";
    $("#logBox").textContent = "";
    hide($("#logBox"));
    $("#logToggle").textContent = t("showLog");
    const bar = $(".progress");
    bar.setAttribute("aria-valuenow", "0");
  }

  function startPolling() {
    stopPolling();
    state.pollTimer = window.setInterval(pollJob, 1000);
    pollJob();
  }

  function stopPolling() {
    if (state.pollTimer) {
      window.clearInterval(state.pollTimer);
      state.pollTimer = null;
    }
  }

  async function pollJob() {
    if (!state.jobId) return;
    try {
      const data = await api(`/api/jobs/${state.jobId}`);
      state.lastJobData = data;
      renderProgress(data);
      if (data.status === "done") {
        stopPolling();
        state.jobId = null;
        state.result = data.result;
        hide($("#progressCard"));
        hide($("#selectCard"));
        renderResult();
        show($("#resultCard"));
        $("#resultCard").scrollIntoView({ behavior: "smooth", block: "start" });
        updateStartState();
      } else if (data.status === "error" || data.status === "cancelled") {
        stopPolling();
        state.jobId = null;
        hide($("#progressCard"));
        show($("#selectCard"));
        showError({
          code: data.error_code || "unknown",
          message: data.error || "",
          log: data.log || [],
        });
        updateStartState();
      }
    } catch (error) {
      stopPolling();
      state.jobId = null;
      hide($("#progressCard"));
      show($("#selectCard"));
      showError(error);
      updateStartState();
    }
  }

  function renderProgress(data) {
    if (!data) return;
    const stageKey = `stage_${data.stage || "queued"}`;
    $("#stageLabel").textContent =
      I18N[state.lang] && I18N[state.lang][stageKey] ? t(stageKey) : t("stage_queued");
    const progress = Math.max(0, Math.min(100, Number(data.progress) || 0));
    const percent = `${Math.round(progress)}%`;
    $("#progressPct").textContent = percent;
    $("#progressBar").style.width = percent;
    $(".progress").setAttribute("aria-valuenow", String(Math.round(progress)));
    $("#stageMessage").textContent = data.message || "";
    const elapsed =
      data.elapsed && data.elapsed > 0
        ? data.elapsed
        : state.startedAt
          ? (Date.now() - state.startedAt) / 1000
          : 0;
    $("#elapsed").textContent = formatDuration(elapsed);
    if (state.logOpen) {
      const log = (data.log || []).join("\n");
      const box = $("#logBox");
      box.textContent = log;
      box.scrollTop = box.scrollHeight;
    }
  }

  async function cancelTest() {
    if (!state.jobId) return;
    $("#cancelBtn").disabled = true;
    try {
      await api(`/api/jobs/${state.jobId}/cancel`, { method: "POST" });
    } catch (error) {
      showError(error);
    } finally {
      $("#cancelBtn").disabled = false;
    }
  }

  function showError(error) {
    state.error = error.message || "";
    state.errorCode = error.code || "unknown";
    state.errorLog = error.log || [];
    hide($("#progressCard"));
    hide($("#resultCard"));
    hide($("#errorCard"));
    $("#errorText").textContent = errorMessage(state.errorCode);
    $("#errorSuggest").textContent = errorSuggestion(state.errorCode);
    const detail = [state.error, ...state.errorLog.slice(-20)].filter(Boolean).join("\n");
    $("#errorDetail").textContent = detail || "—";
    $("#errorDetails").open = false;
    if (state.configs.length) show($("#selectCard"));
    const probeLike = ["probe_failed", "resolve_failed", "timeout", "xray_failed"].includes(
      state.errorCode,
    );
    $("#anotherBtn").classList.toggle("hidden", !probeLike || !state.configs.length);
    show($("#errorCard"));
    $("#errorCard").scrollIntoView({ behavior: "smooth", block: "center" });
  }

  function renderResult() {
    const result = state.result;
    if (!result) return;
    const ratio = Number(result.ratio) || 0;
    const coef = $("#coef");
    coef.textContent = formatRatio(ratio);
    coef.classList.remove("warn", "bad");
    if (ratio > 1.5) coef.classList.add("bad");
    else if (ratio > 1.1) coef.classList.add("warn");

    const verdict = $("#verdict");
    const claimed = result.claimed_ratio;
    if (claimed !== null && claimed !== undefined) {
      const matches = Boolean(result.claimed_matches);
      let key = "verdictMatch";
      let cls = "good";
      if (!matches) {
        if (ratio > Number(claimed)) {
          key = "verdictHigher";
          cls = "bad";
        } else {
          key = "verdictLower";
          cls = "good";
        }
      }
      verdict.textContent = t(key);
      verdict.className = `badge ${cls}`;
      show(verdict);
    } else {
      hide(verdict);
    }

    $("#sentence").textContent = t("sentence", { n: ratio.toFixed(2) });
    $("#statDownloaded").textContent = formatBytes(result.downloaded_bytes);
    $("#statDeducted").textContent = formatBytes(result.delta_total);
    $("#statClaimed").textContent =
      claimed === null || claimed === undefined ? t("notClaimed") : formatRatio(claimed);

    const remaining = result.remaining_bytes;
    const remainingEl = $("#statRemaining");
    remainingEl.replaceChildren();
    if (remaining === null || remaining === undefined) {
      remainingEl.textContent = "—";
    } else {
      remainingEl.textContent = formatBytes(remaining);
      if (ratio > 0) {
        const sub = document.createElement("span");
        sub.className = "stat-sub";
        sub.textContent = t("realEquivalent", { v: formatBytes(remaining / ratio) });
        remainingEl.append(sub);
      }
    }

    const meta = $("#resultMeta");
    meta.replaceChildren();
    const rows = [];
    if (result.config) rows.push([t("configUsed"), result.config]);
    if (result.core) {
      const version = result.xray_version ? ` — ${result.xray_version}` : "";
      rows.push([t("coreUsed"), `${result.core}${version}`]);
    }
    if (result.duration_seconds) rows.push([t("duration"), formatDuration(result.duration_seconds)]);
    rows.forEach(([label, value]) => {
      const row = document.createElement("span");
      row.className = "meta-row";
      const b = document.createElement("b");
      b.textContent = `${label}:`;
      const span = document.createElement("span");
      span.textContent = value;
      row.append(b, span);
      meta.append(row);
    });
  }

  function shareCaption(result) {
    const ratio = Number(result.ratio) || 0;
    return [
      `${t("shareCaptionTitle")} — ${formatRatio(ratio)}`,
      t("sentence", { n: ratio.toFixed(2) }),
      SITE_URL,
    ].join("\n");
  }

  function roundRect(ctx, x, y, w, h, r) {
    const radius = Math.min(r, w / 2, h / 2);
    ctx.beginPath();
    if (typeof ctx.roundRect === "function") {
      ctx.roundRect(x, y, w, h, radius);
      return;
    }
    ctx.moveTo(x + radius, y);
    ctx.arcTo(x + w, y, x + w, y + h, radius);
    ctx.arcTo(x + w, y + h, x, y + h, radius);
    ctx.arcTo(x, y + h, x, y, radius);
    ctx.arcTo(x, y, x + w, y, radius);
    ctx.closePath();
  }

  function setCanvasFont(ctx, weight, size, mono = false) {
    const family = mono
      ? "ui-monospace, Menlo, Consolas, monospace"
      : '"Vazirmatn", "Segoe UI", Tahoma, "Noto Sans Arabic", system-ui, -apple-system, sans-serif';
    ctx.font = `${weight} ${size}px ${family}`;
  }

  function fitText(ctx, text, maxWidth) {
    let value = String(text);
    if (ctx.measureText(value).width <= maxWidth) return value;
    while (value.length > 1 && ctx.measureText(`${value}…`).width > maxWidth) {
      value = value.slice(0, -1);
    }
    return `${value}…`;
  }

  function wrapLines(ctx, text, maxWidth) {
    const words = String(text).split(/\s+/).filter(Boolean);
    const lines = [];
    let line = "";
    for (const word of words) {
      const candidate = line ? `${line} ${word}` : word;
      if (ctx.measureText(candidate).width <= maxWidth) {
        line = candidate;
        continue;
      }
      if (line) lines.push(line);
      if (ctx.measureText(word).width <= maxWidth) {
        line = word;
        continue;
      }
      let chunk = "";
      for (const char of word) {
        if (chunk && ctx.measureText(chunk + char).width > maxWidth) {
          lines.push(chunk);
          chunk = char;
        } else {
          chunk += char;
        }
      }
      line = chunk;
    }
    if (line) lines.push(line);
    return lines;
  }

  function drawCanvasText(ctx, text, x, y, options = {}) {
    const {
      size = 28,
      weight = 400,
      color = "#0f172a",
      align = "left",
      rtl = false,
      mono = false,
    } = options;
    setCanvasFont(ctx, weight, size, mono);
    ctx.fillStyle = color;
    ctx.textAlign = align;
    ctx.textBaseline = "alphabetic";
    ctx.direction = rtl ? "rtl" : "ltr";
    ctx.fillText(text, x, y);
  }

  function shareColors(theme) {
    if (theme === "dark") {
      return {
        bg1: "#050810",
        bg2: "#070c17",
        card: "#0d1424",
        line: "rgba(148, 163, 184, 0.18)",
        text: "#e8effc",
        muted: "#91a3c1",
        muted2: "#6f819f",
        accent: "#22d3ee",
        good: "#34d399",
        violet: "#a78bfa",
        warn: "#fbbf24",
        danger: "#fb7185",
        goodSoft: "rgba(52, 211, 153, 0.1)",
        goodBorder: "rgba(52, 211, 153, 0.3)",
        dangerSoft: "rgba(251, 113, 133, 0.1)",
        dangerBorder: "rgba(251, 113, 133, 0.3)",
        chip: "rgba(34, 211, 238, 0.08)",
        chipBorder: "rgba(34, 211, 238, 0.22)",
        glow1: "rgba(34, 211, 238, 0.16)",
        glow2: "rgba(167, 139, 250, 0.12)",
        primary: "#041018",
      };
    }
    return {
      bg1: "#f7faff",
      bg2: "#eef4fb",
      card: "#ffffff",
      line: "rgba(15, 23, 42, 0.1)",
      text: "#0f172a",
      muted: "#52627a",
      muted2: "#7b8aa3",
      accent: "#0891b2",
      good: "#059669",
      violet: "#7c3aed",
      warn: "#d97706",
      danger: "#e11d48",
      goodSoft: "rgba(5, 150, 105, 0.08)",
      goodBorder: "rgba(5, 150, 105, 0.25)",
      dangerSoft: "rgba(225, 29, 72, 0.07)",
      dangerBorder: "rgba(225, 29, 72, 0.25)",
      chip: "rgba(8, 145, 178, 0.08)",
      chipBorder: "rgba(8, 145, 178, 0.22)",
      glow1: "rgba(6, 182, 212, 0.12)",
      glow2: "rgba(139, 92, 246, 0.08)",
      primary: "#041018",
    };
  }

  async function renderShareImage(result) {
    try {
      await document.fonts.ready;
    } catch (_) {
      /* ignore */
    }
    const W = 1080;
    const H = 1350;
    const pad = 72;
    const canvas = document.createElement("canvas");
    canvas.width = W;
    canvas.height = H;
    const ctx = canvas.getContext("2d");
    const rtl = state.lang === "fa";
    const theme =
      document.documentElement.getAttribute("data-theme") === "dark" ? "dark" : "light";
    const C = shareColors(theme);
    const ratio = Number(result.ratio) || 0;
    const ratioColor = ratio > 1.5 ? C.danger : ratio > 1.1 ? C.warn : C.good;

    const bg = ctx.createLinearGradient(0, 0, 0, H);
    bg.addColorStop(0, C.bg1);
    bg.addColorStop(1, C.bg2);
    ctx.fillStyle = bg;
    ctx.fillRect(0, 0, W, H);

    const glow1 = ctx.createRadialGradient(W * 0.15, -40, 0, W * 0.15, -40, 720);
    glow1.addColorStop(0, C.glow1);
    glow1.addColorStop(1, "rgba(0,0,0,0)");
    ctx.fillStyle = glow1;
    ctx.fillRect(0, 0, W, H);

    const glow2 = ctx.createRadialGradient(W * 0.95, 180, 0, W * 0.95, 180, 620);
    glow2.addColorStop(0, C.glow2);
    glow2.addColorStop(1, "rgba(0,0,0,0)");
    ctx.fillStyle = glow2;
    ctx.fillRect(0, 0, W, H);

    const markX = pad;
    const markY = 62;
    const markSize = 58;
    const markGrad = ctx.createLinearGradient(markX, markY, markX + markSize, markY + markSize);
    markGrad.addColorStop(0, C.accent);
    markGrad.addColorStop(1, C.violet);
    roundRect(ctx, markX, markY, markSize, markSize, 16);
    ctx.fillStyle = markGrad;
    ctx.fill();
    ctx.strokeStyle = C.primary;
    ctx.lineWidth = 4;
    ctx.lineCap = "round";
    ctx.lineJoin = "round";
    ctx.beginPath();
    ctx.moveTo(markX + 13, markY + 40);
    ctx.lineTo(markX + 24, markY + 17);
    ctx.lineTo(markX + 34, markY + 34);
    ctx.lineTo(markX + 40, markY + 24);
    ctx.lineTo(markX + 46, markY + 40);
    ctx.stroke();

    setCanvasFont(ctx, 900, 40);
    const ratioWidth = ctx.measureText("Ratio").width;
    drawCanvasText(ctx, "Ratio", markX + markSize + 18, 106, {
      size: 40,
      weight: 900,
      color: C.text,
    });
    drawCanvasText(ctx, "Test", markX + markSize + 18 + ratioWidth, 106, {
      size: 40,
      weight: 900,
      color: C.accent,
    });
    drawCanvasText(ctx, SITE_URL, W - pad, 104, {
      size: 22,
      weight: 600,
      color: C.muted2,
      align: "right",
      mono: true,
    });

    ctx.strokeStyle = C.line;
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(pad, 162);
    ctx.lineTo(W - pad, 162);
    ctx.stroke();

    const eyebrowText = t("resultEyebrow").toUpperCase();
    setCanvasFont(ctx, 700, 22);
    const eyebrowWidth = ctx.measureText(eyebrowText).width;
    const pillW = eyebrowWidth + 46;
    const pillH = 46;
    const pillX = rtl ? W - pad - pillW : pad;
    const pillY = 208;
    roundRect(ctx, pillX, pillY, pillW, pillH, 23);
    ctx.fillStyle = C.chip;
    ctx.fill();
    ctx.strokeStyle = C.chipBorder;
    ctx.lineWidth = 2;
    ctx.stroke();
    drawCanvasText(ctx, eyebrowText, rtl ? pillX + pillW - 23 : pillX + 23, pillY + 31, {
      size: 22,
      weight: 700,
      color: C.accent,
      align: rtl ? "right" : "left",
    });

    drawCanvasText(ctx, formatRatio(ratio), rtl ? W - pad : pad, 430, {
      size: 190,
      weight: 900,
      color: ratioColor,
      align: rtl ? "right" : "left",
    });

    let verdictY = 470;
    const claimed = result.claimed_ratio;
    if (claimed !== null && claimed !== undefined) {
      const matches = Boolean(result.claimed_matches);
      let key = "verdictMatch";
      let tone = C.good;
      let soft = C.goodSoft;
      let border = C.goodBorder;
      if (!matches) {
        if (ratio > Number(claimed)) {
          key = "verdictHigher";
          tone = C.danger;
          soft = C.dangerSoft;
          border = C.dangerBorder;
        } else {
          key = "verdictLower";
        }
      }
      const verdictText = t(key);
      setCanvasFont(ctx, 700, 24);
      const verdictWidth = ctx.measureText(verdictText).width;
      const verdictW = verdictWidth + 44;
      const verdictH = 44;
      const verdictX = rtl ? W - pad - verdictW : pad;
      roundRect(ctx, verdictX, verdictY, verdictW, verdictH, 22);
      ctx.fillStyle = soft;
      ctx.fill();
      ctx.strokeStyle = border;
      ctx.lineWidth = 2;
      ctx.stroke();
      drawCanvasText(
        ctx,
        verdictText,
        rtl ? verdictX + verdictW - 22 : verdictX + 22,
        verdictY + 30,
        {
          size: 24,
          weight: 700,
          color: tone,
          align: rtl ? "right" : "left",
          rtl: true,
        },
      );
      verdictY += 64;
    }

    const sentence = t("sentence", { n: ratio.toFixed(2) });
    const maxTextWidth = W - pad * 2;
    setCanvasFont(ctx, 500, 33);
    const sentenceLines = wrapLines(ctx, sentence, maxTextWidth);
    let textY = Math.max(555, verdictY + 34);
    sentenceLines.forEach((line) => {
      drawCanvasText(ctx, line, rtl ? W - pad : pad, textY, {
        size: 33,
        weight: 500,
        color: C.muted,
        align: rtl ? "right" : "left",
        rtl,
      });
      textY += 52;
    });

    const statsTop = Math.max(690, textY + 28);
    const statGap = 20;
    const statW = (W - pad * 2 - statGap) / 2;
    const statH = 150;
    const remaining = result.remaining_bytes;
    const stats = [
      { label: t("downloaded"), value: formatBytes(result.downloaded_bytes) },
      { label: t("deducted"), value: formatBytes(result.delta_total) },
      {
        label: t("claimed"),
        value:
          claimed === null || claimed === undefined ? t("notClaimed") : formatRatio(claimed),
      },
      {
        label: t("remaining"),
        value: remaining === null || remaining === undefined ? "—" : formatBytes(remaining),
        sub:
          remaining !== null && remaining !== undefined && ratio > 0
            ? t("realEquivalent", { v: formatBytes(remaining / ratio) })
            : "",
      },
    ];
    stats.forEach((stat, index) => {
      const row = Math.floor(index / 2);
      const col = index % 2;
      const x = rtl
        ? W - pad - statW - col * (statW + statGap)
        : pad + col * (statW + statGap);
      const y = statsTop + row * (statH + statGap);
      roundRect(ctx, x, y, statW, statH, 24);
      ctx.fillStyle = C.card;
      ctx.fill();
      ctx.strokeStyle = C.line;
      ctx.lineWidth = 2;
      ctx.stroke();
      const textX = rtl ? x + statW - 26 : x + 26;
      const align = rtl ? "right" : "left";
      drawCanvasText(ctx, stat.label, textX, y + 44, {
        size: 23,
        weight: 600,
        color: C.muted2,
        align,
        rtl,
      });
      drawCanvasText(ctx, stat.value, textX, y + (stat.sub ? 98 : 108), {
        size: 36,
        weight: 800,
        color: C.text,
        align,
      });
      if (stat.sub) {
        drawCanvasText(ctx, stat.sub, textX, y + 130, {
          size: 20,
          weight: 500,
          color: C.muted2,
          align,
          rtl,
        });
      }
    });

    const metaY = statsTop + statH * 2 + statGap + 44;
    const metaLines = [];
    if (result.config) metaLines.push(`${t("configUsed")}: ${result.config}`);
    if (result.core) {
      const version = result.xray_version ? ` — ${result.xray_version}` : "";
      metaLines.push(`${t("coreUsed")}: ${result.core}${version}`);
    }
    if (result.duration_seconds) {
      metaLines.push(`${t("duration")}: ${formatDuration(result.duration_seconds)}`);
    }
    if (result.tested_at) {
      metaLines.push(`${t("resultEyebrow")}: ${formatDate(result.tested_at)}`);
    }
    setCanvasFont(ctx, 500, 23);
    let metaLineY = metaY;
    metaLines.slice(0, 2).forEach((line) => {
      const fitted = fitText(ctx, line, maxTextWidth);
      drawCanvasText(ctx, fitted, rtl ? W - pad : pad, metaLineY, {
        size: 23,
        weight: 500,
        color: C.muted2,
        align: rtl ? "right" : "left",
        rtl,
      });
      metaLineY += 36;
    });

    const footerLineY = H - 178;
    ctx.strokeStyle = C.line;
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(pad, footerLineY);
    ctx.lineTo(W - pad, footerLineY);
    ctx.stroke();

    drawCanvasText(ctx, SITE_URL, W / 2, footerLineY + 62, {
      size: 36,
      weight: 800,
      color: C.accent,
      align: "center",
      mono: true,
    });
    drawCanvasText(ctx, t("shareFooter"), W / 2, footerLineY + 108, {
      size: 23,
      weight: 500,
      color: C.muted2,
      align: "center",
      rtl,
    });

    const bar = ctx.createLinearGradient(pad, 0, W - pad, 0);
    bar.addColorStop(0, C.accent);
    bar.addColorStop(0.5, C.good);
    bar.addColorStop(1, C.violet);
    ctx.fillStyle = bar;
    ctx.fillRect(pad, H - 12, W - pad * 2, 6);

    return canvasToBlob(canvas);
  }

  function canvasToBlob(canvas) {
    return new Promise((resolve, reject) => {
      canvas.toBlob((blob) => {
        if (blob) resolve(blob);
        else reject(new Error("Canvas export failed"));
      }, "image/png");
    });
  }

  function downloadBlob(blob, filename) {
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.setTimeout(() => URL.revokeObjectURL(url), 4000);
  }

  async function copyText(text) {
    try {
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
        return true;
      }
    } catch (_) {
      /* fall through */
    }
    try {
      const area = document.createElement("textarea");
      area.value = text;
      area.setAttribute("readonly", "");
      area.style.position = "fixed";
      area.style.opacity = "0";
      document.body.appendChild(area);
      area.select();
      const copied = document.execCommand("copy");
      area.remove();
      return copied;
    } catch (_) {
      return false;
    }
  }

  async function shareResult() {
    if (!state.result) return;
    const button = $("#shareBtn");
    const label = $(".btn-label", button);
    if (button) button.disabled = true;
    if (label) label.textContent = t("sharePreparing");
    try {
      const blob = await renderShareImage(state.result);
      const file = new File([blob], "ratiotest-result.png", { type: "image/png" });
      const caption = shareCaption(state.result);
      let shared = false;
      if (navigator.canShare) {
        try {
          if (navigator.canShare({ files: [file] })) {
            await navigator.share({
              files: [file],
              text: caption,
              title: t("pageTitle"),
            });
            shared = true;
          }
        } catch (error) {
          if (error && error.name === "AbortError") {
            shared = true;
          }
        }
      }
      if (!shared) {
        downloadBlob(blob, "ratiotest-result.png");
        const copied = await copyText(caption);
        toast(copied ? t("shareFallback") : t("shareDownloaded"), 6000);
      }
    } catch (error) {
      console.error(error);
      toast(t("shareFailed"));
    } finally {
      if (button) button.disabled = false;
      if (label) label.textContent = t("shareResult");
    }
  }

  function testAnother() {
    hide($("#resultCard"));
    hide($("#errorCard"));
    hide($("#progressCard"));
    show($("#selectCard"));
    $("#agree").checked = true;
    state.jobId = null;
    state.result = null;
    state.error = null;
    state.claimedTouched = false;
    updateStartState();
    $("#selectCard").scrollIntoView({ behavior: "smooth", block: "start" });
  }

  function retry() {
    hide($("#errorCard"));
    if (state.selected && state.configs.length) {
      if ($("#agree").checked) {
        startTest();
        return;
      }
      show($("#selectCard"));
      updateStartState();
      $("#selectCard").scrollIntoView({ behavior: "smooth", block: "start" });
      return;
    }
    if (state.configs.length) {
      show($("#selectCard"));
      updateStartState();
      $("#selectCard").scrollIntoView({ behavior: "smooth", block: "start" });
    } else {
      $("#subUrl").focus();
      $("#subUrl").scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }

  function setLang(lang) {
    state.lang = lang === "en" ? "en" : "fa";
    try {
      localStorage.setItem("ratio-lang", state.lang);
    } catch (_) {
      /* ignore */
    }
    applyI18n();
    renderUsage();
    renderConfigs();
    renderClaim();
    updateStartState();
    if (state.result) renderResult();
    if (state.lastJobData && !$("#progressCard").classList.contains("hidden")) {
      renderProgress(state.lastJobData);
    }
    if (state.errorCode && !$("#errorCard").classList.contains("hidden")) {
      showError({ code: state.errorCode, message: state.error, log: state.errorLog });
    }
  }

  function init() {
    applyI18n();
    applyTheme(document.documentElement.getAttribute("data-theme"), false);
    setSize(10);

    $("#themeBtn").addEventListener("click", toggleTheme);
    $("#langBtn").addEventListener("click", () => {
      setLang(state.lang === "fa" ? "en" : "fa");
    });
    $("#inspectBtn").addEventListener("click", inspect);
    $("#subUrl").addEventListener("keydown", (event) => {
      if (event.key === "Enter") {
        event.preventDefault();
        inspect();
      }
    });
    $("#sizeGroup").addEventListener("click", (event) => {
      const button = event.target.closest(".seg");
      if (button) setSize(button.dataset.size);
    });
    $("#claimed").addEventListener("input", () => {
      state.claimedTouched = true;
    });
    $("#agree").addEventListener("change", updateStartState);
    $("#startBtn").addEventListener("click", startTest);
    $("#cancelBtn").addEventListener("click", cancelTest);
    $("#logToggle").addEventListener("click", () => {
      state.logOpen = !state.logOpen;
      updateLogToggle();
      const box = $("#logBox");
      if (state.logOpen) {
        show(box);
        if (state.lastJobData) {
          box.textContent = (state.lastJobData.log || []).join("\n");
          box.scrollTop = box.scrollHeight;
        }
      } else {
        hide(box);
      }
    });
    $("#againBtn").addEventListener("click", testAnother);
    $("#shareBtn").addEventListener("click", shareResult);
    $("#retryBtn").addEventListener("click", retry);
    $("#anotherBtn").addEventListener("click", testAnother);
  }

  init();
})();
