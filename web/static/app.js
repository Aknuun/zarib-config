(() => {
  "use strict";

  const I18N = {
    fa: {
      pageTitle: "محاسبه ضریب واقعی ترافیک کانفیگ | RatioTest",
      eyebrow: "شفافیت حجم، بدون حدس",
      heroTitle1: "محاسبه ضریب واقعی",
      heroTitle2: "ترافیک کانفیگ",
      heroSub: "۵ مگابایت تست، بدون ذخیره لینک.",
      privacyNote: "طبق مستندات گیت‌هاب، هیچ کانفیگی نگه‌داشته یا ذخیره نمی‌شود.",
      chip1: "مگابایت تست",
      chip2: "بدون ذخیره لینک",
      chip3: "مقایسه با ادعای پنل",
      subTitle: "لینک سابسکریپشن",
      statusReady: "آماده",
      subLabel: "لینک سابسکریپشن",
      loadConfigs: "دریافت کانفیگ‌ها",
      pasteSubLink: "جاگذاری لینک ساب",
      pasteHint: "لینک ساب را در کادر بچسبان (Ctrl+V).",
      subLimit: "حداکثر ۵ لینک ساب می‌توانی اضافه کنی.",
      removeSub: "حذف لینک",
      subLoading: "در حال خواندن…",
      subConfigs: "{n} کانفیگ",
      configsToTest: "تعداد کانفیگ تست",
      configsHint: "برای هر ساب، این تعداد کانفیگ اول انتخاب می‌شود. کانفیگ‌های یک ساب به‌ترتیب و ساب‌ها هم‌زمان تست می‌شوند.",
      perConfigResults: "نتیجه هر کانفیگ",
      sentenceAvg: "میانگین ضریب {n} کانفیگ: {r}×",
      err_no_selection: "حداقل یک کانفیگ برای تست انتخاب کن.",
      resultError: "خطا",
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
      chooseAnother: "کانفیگ دیگری از همین ساب",
      claimOptional: "اختیاری",
      claimDetected: "ادعای این کانفیگ: {n}",
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
      usageSlowHint: "بعضی ساب‌ها دیر نتیجه می‌دهند؛ این ربطی به ماشین‌حساب ضریب ندارد.",
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
      coefMeasured: "اندازه‌گیری واقعی: {r}",
      noRatioClose: "کاملاً بدون ضریب است",
      noRatioExact: "این کانفیگ یا ساب ضریب ندارد",
      background: "اصلاح پس‌زمینه",
      raw: "خام",
      duration: "مدت تست",
      configUsed: "کانفیگ",
      testAnother: "کانفیگ دیگری از همین ساب",
      shareResult: "اشتراک‌گذاری",
      sharePreparing: "در حال ساخت تصویر…",
      shareFallback: "تصویر نتیجه دانلود شد و متن اشتراک‌گذاری کپی شد. آن را در تلگرام بفرست.",
      shareDownloaded: "تصویر نتیجه دانلود شد. متن اشتراک‌گذاری را دستی کپی کن.",
      shareFailed: "ساخت یا اشتراک‌گذاری تصویر ممکن نشد.",
      shareCaptionTitle: "نتیجه محاسبه ضریب ترافیک کانفیگ من",
      shareFooter: "محاسبه ضریب واقعی ترافیک کانفیگ",
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
      pageTitle: "Config Traffic Ratio Calculator | RatioTest",
      eyebrow: "Transparent data accounting",
      heroTitle1: "Calculate the real",
      heroTitle2: "config traffic ratio",
      heroSub: "A 5 MB test. Link never stored.",
      privacyNote: "Per our GitHub docs, no config is retained or stored.",
      chip1: "MB test",
      chip2: "Link never stored",
      chip3: "Compares with the claim",
      subTitle: "Subscription link",
      statusReady: "Ready",
      subLabel: "Subscription URL",
      loadConfigs: "Load configs",
      pasteSubLink: "Paste sub link",
      pasteHint: "Paste the subscription link in the box (Ctrl+V).",
      subLimit: "You can add up to 5 subscription links.",
      removeSub: "Remove link",
      subLoading: "Loading…",
      subConfigs: "{n} configs",
      configsToTest: "Configs to test",
      configsHint: "The first N configs of each subscription are selected. Configs within a subscription run sequentially; subscriptions run in parallel.",
      perConfigResults: "Result per config",
      sentenceAvg: "Average across {n} configs: {r}×",
      err_no_selection: "Select at least one config to test.",
      resultError: "Error",
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
      chooseAnother: "Another config from this sub",
      claimOptional: "Optional",
      claimDetected: "This config claims: {n}",
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
      usageSlowHint: "Some subscriptions report usage slowly; this is not caused by the ratio calculator.",
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
      coefMeasured: "Measured: {r}",
      noRatioClose: "Completely without a coefficient",
      noRatioExact: "This config or sub has no coefficient",
      background: "Background correction",
      raw: "raw",
      duration: "Duration",
      configUsed: "Config",
      testAnother: "Another config from this sub",
      shareResult: "Share",
      sharePreparing: "Creating image…",
      shareFallback: "The result image was downloaded and the share text was copied. Attach it in Telegram.",
      shareDownloaded: "The result image was downloaded. Copy the share text manually.",
      shareFailed: "Could not create or share the result image.",
      shareCaptionTitle: "My config traffic ratio result",
      shareFooter: "Real config traffic ratio calculator",
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
    subs: [],
    testCount: 1,
    size: 5,
    claimedTouched: false,
    running: false,
    activeJobs: new Set(),
    progressTimer: null,
    displayPct: 0,
    targetPct: 0,
    startedAt: 0,
    usageWaitStarted: 0,
    lastJobData: null,
    logOpen: false,
    pasteTimer: null,
    results: [],
    completed: 0,
    total: 0,
    result: null,
    error: null,
    errorCode: null,
    errorLog: [],
  };
  let subSeq = 0;

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

  function ratioIsNo(value) {
    const n = Number(value) || 0;
    return n > 0 && n < 1.1;
  }

  function ratioIsUnit(value) {
    return Math.round((Number(value) || 0) * 100) === 100;
  }

  function ratioBig(value) {
    return ratioIsNo(value) ? "1.00x" : formatRatio(value);
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

  function maxConfigs() {
    return Math.max(1, Math.min(5, Number(state.testCount) || 1));
  }

  function defaultSelection(sub) {
    return sub.configs.slice(0, maxConfigs()).map((config) => config.index);
  }

  function subHost(url) {
    try {
      return new URL(url).host || url;
    } catch (_) {
      return url;
    }
  }

  async function addSub(raw) {
    if (state.running) return;
    clearInputError();
    const normalized = normalizeUrl(raw);
    if (!normalized) {
      showInputError(errorMessage("invalid_url"));
      return;
    }
    if (state.subs.length >= 5) {
      showInputError(t("subLimit"));
      return;
    }
    if (state.subs.some((sub) => sub.url === normalized)) {
      $("#subUrl").value = "";
      return;
    }
    const sub = {
      id: `sub-${++subSeq}`,
      url: normalized,
      host: subHost(normalized),
      loading: true,
      error: null,
      configs: [],
      count: 0,
      selected: [],
    };
    state.subs.push(sub);
    $("#subUrl").value = "";
    show($("#selectCard"));
    hide($("#resultCard"));
    hide($("#errorCard"));
    hide($("#progressCard"));
    renderSubs();
    updateStartState();
    await inspectSub(sub);
  }

  function addSubFromInput() {
    const value = $("#subUrl").value.trim();
    if (!value) {
      pasteFromClipboard();
      return;
    }
    const parts = value.split(/\s+/).filter(Boolean);
    if (parts.length > 1) {
      $("#subUrl").value = "";
      parts.forEach((part) => addSub(part));
      return;
    }
    addSub(value);
  }

  function scheduleAutoAddFromPaste() {
    if (state.pasteTimer) window.clearTimeout(state.pasteTimer);
    state.pasteTimer = window.setTimeout(() => {
      state.pasteTimer = null;
      if ($("#subUrl").value.trim()) addSubFromInput();
    }, 120);
  }

  function handlePaste(event) {
    const text = event.clipboardData ? event.clipboardData.getData("text") : "";
    const parts = text.split(/\s+/).filter(Boolean);
    if (parts.length > 1) {
      event.preventDefault();
      $("#subUrl").value = "";
      parts.forEach((part) => addSub(part));
      return;
    }
    scheduleAutoAddFromPaste();
  }

  async function pasteFromClipboard() {
    try {
      if (navigator.clipboard && window.isSecureContext) {
        const text = await navigator.clipboard.readText();
        if (text && text.trim()) {
          $("#subUrl").value = text.trim();
          addSub(text.trim());
          return;
        }
      }
    } catch (_) {
      /* ignore */
    }
    $("#subUrl").focus();
    toast(t("pasteHint"));
  }

  async function inspectSub(sub) {
    sub.loading = true;
    sub.error = null;
    renderSubs();
    updateStartState();
    try {
      const data = await api("/api/inspect", {
        method: "POST",
        body: JSON.stringify({ sub_url: sub.url }),
      });
      sub.configs = data.configs || [];
      sub.count = data.count || sub.configs.length;
      sub.selected = defaultSelection(sub);
    } catch (error) {
      sub.error = error;
      sub.configs = [];
      sub.count = 0;
      sub.selected = [];
    } finally {
      sub.loading = false;
      renderSubs();
      renderClaim();
      updateStartState();
      if (state.subs.some((item) => item.configs.length)) {
        $("#selectCard").scrollIntoView({ behavior: "smooth", block: "start" });
      }
    }
  }

  function removeSub(id) {
    if (state.running) return;
    state.subs = state.subs.filter((sub) => sub.id !== id);
    renderSubs();
    renderClaim();
    updateStartState();
    if (!state.subs.length) hide($("#selectCard"));
  }

  function renderSubs() {
    const container = $("#subsContainer");
    const scrolls = new Map();
    container.querySelectorAll(".sub-group").forEach((group) => {
      const list = group.querySelector(".config-list");
      if (list && group.dataset.subId) scrolls.set(group.dataset.subId, list.scrollTop);
    });
    container.replaceChildren();
    let totalConfigs = 0;
    state.subs.forEach((sub) => {
      totalConfigs += sub.count || sub.configs.length;
      const group = document.createElement("section");
      group.className = "sub-group";
      group.dataset.subId = sub.id;

      const head = document.createElement("div");
      head.className = "sub-head";
      const title = document.createElement("div");
      title.className = "sub-title";
      const host = document.createElement("strong");
      host.textContent = sub.host || sub.url;
      const status = document.createElement("span");
      status.className = "sub-status";
      if (sub.loading) status.textContent = t("subLoading");
      else if (sub.error) status.textContent = errorMessage(sub.error.code);
      else status.textContent = t("subConfigs", { n: sub.count || sub.configs.length });
      title.append(host, status);
      const remove = document.createElement("button");
      remove.type = "button";
      remove.className = "ghost-btn sub-remove";
      remove.textContent = t("removeSub");
      remove.disabled = state.running;
      remove.addEventListener("click", () => removeSub(sub.id));
      head.append(title, remove);
      group.append(head);

      if (sub.loading) {
        const loading = document.createElement("p");
        loading.className = "hint small";
        loading.textContent = t("subLoading");
        group.append(loading);
      } else if (sub.error) {
        const error = document.createElement("p");
        error.className = "inline-error";
        error.textContent = errorMessage(sub.error.code);
        group.append(error);
      } else if (!sub.configs.length) {
        const empty = document.createElement("p");
        empty.className = "hint small";
        empty.textContent = t("err_no_configs");
        group.append(empty);
      } else {
        const list = document.createElement("div");
        list.className = "config-list";
        sub.configs.forEach((config) => {
          const item = document.createElement("label");
          item.className = "config-item";
          item.dataset.index = String(config.index);
          if (maxConfigs() > 1) item.classList.add("multi");
          if (sub.selected.includes(config.index)) item.classList.add("selected");

          const check = document.createElement("input");
          check.type = "checkbox";
          check.className = "config-radio";
          check.checked = sub.selected.includes(config.index);
          check.addEventListener("change", () => toggleConfig(sub, config.index));

          const proto = document.createElement("span");
          proto.className = "proto";
          proto.textContent = config.protocol || "json";
          const name = document.createElement("span");
          name.className = "config-name";
          name.textContent = config.remark || config.target || `#${config.index}`;
          item.title = config.target || "";
          item.append(check, proto, name);
          list.append(item);
        });
        group.append(list);
      }
      container.append(group);
    });

    container.querySelectorAll(".sub-group").forEach((group) => {
      const list = group.querySelector(".config-list");
      if (list && group.dataset.subId && scrolls.has(group.dataset.subId)) {
        list.scrollTop = scrolls.get(group.dataset.subId);
      }
    });

    const count = $("#configCount");
    count.textContent = t("configCount", { n: totalConfigs });
  }

  function syncSelection(sub) {
    const group = $("#subsContainer").querySelector(`[data-sub-id="${sub.id}"]`);
    if (!group) return;
    group.querySelectorAll(".config-item").forEach((item) => {
      const index = Number(item.dataset.index);
      const selected = sub.selected.includes(index);
      item.classList.toggle("selected", selected);
      const input = item.querySelector(".config-radio");
      if (input) input.checked = selected;
    });
  }

  function toggleConfig(sub, index) {
    if (state.running) return;
    if (sub.selected.includes(index)) {
      sub.selected = sub.selected.filter((value) => value !== index);
    } else {
      const limit = maxConfigs();
      const next = sub.selected.filter((value) => value !== index);
      next.push(index);
      while (next.length > limit) next.shift();
      sub.selected = next;
    }
    syncSelection(sub);
    renderClaim();
    updateStartState();
  }

  function selectConfigsForCount() {
    state.subs.forEach((sub) => {
      sub.selected = defaultSelection(sub);
    });
    renderSubs();
    renderClaim();
    updateStartState();
  }

  function selectedConfig() {
    for (const sub of state.subs) {
      const config = sub.configs.find((item) => sub.selected.includes(item.index));
      if (config) return config;
    }
    return null;
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

  function updateLogToggle() {
    $("#logToggle").textContent = state.logOpen ? t("hideLog") : t("showLog");
  }

  function selectedCount() {
    return state.subs.reduce((sum, sub) => sum + sub.selected.length, 0);
  }

  function updateStartState() {
    const hasSelection = selectedCount() > 0;
    $("#startBtn").disabled = !hasSelection || state.running;
    $("#inspectBtn").disabled = state.subs.length >= 5 || state.running;
    $("#testCount").disabled = state.running;
    $("#sizeRange").disabled = state.running;
    $("#coreSelect").disabled = state.running;
  }

  function setSize(size) {
    const value = Math.max(2, Math.min(10, Math.round(Number(size) || 5)));
    state.size = value;
    $("#sizeRange").value = String(value);
    $("#sizeValue").textContent = `${value} MB`;
  }

  async function startTest() {
    if (state.running) return;
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
    const groups = state.subs.filter((sub) => sub.selected.length > 0 && sub.configs.length);
    if (!groups.length) {
      showInputError(t("err_no_selection"));
      return;
    }
    const button = $("#startBtn");
    setLoading(button, true);
    button.disabled = true;
    hide($("#errorCard"));
    hide($("#resultCard"));
    hide($("#selectCard"));
    state.running = true;
    state.claimed = claimed;
    state.results = [];
    state.completed = 0;
    state.total = groups.reduce((sum, sub) => sum + sub.selected.length, 0);
    state.startedAt = Date.now();
    state.lastJobData = null;
    state.logOpen = false;
    state.result = null;
    state.error = null;
    resetProgress();
    show($("#progressCard"));
    $("#progressCard").scrollIntoView({ behavior: "smooth", block: "start" });
    updateStartState();
    try {
      await Promise.all(groups.map((sub) => runSubTests(sub, claimed)));
    } finally {
      state.running = false;
      state.activeJobs.clear();
      setLoading(button, false);
      updateStartState();
    }
    finishTests();
  }

  async function runSubTests(sub, claimed) {
    for (const pick of [...sub.selected]) {
      try {
        const result = await runOne(sub, pick, claimed);
        state.results.push({ sub, pick, result, error: null });
      } catch (error) {
        state.results.push({ sub, pick, result: null, error });
      }
      state.completed += 1;
      updateOverallProgress();
    }
  }

  async function runOne(sub, pick, claimed) {
    const data = await api("/api/jobs", {
      method: "POST",
      body: JSON.stringify({
        sub_url: sub.url,
        pick,
        size_mb: state.size,
        claimed,
        wait: 2,
        drift: 5,
        core: $("#coreSelect").value || "auto",
      }),
    });
    const jobId = data.id;
    state.activeJobs.add(jobId);
    try {
      const job = await waitForJob(jobId);
      return job.result;
    } finally {
      state.activeJobs.delete(jobId);
    }
  }

  function delay(ms) {
    return new Promise((resolve) => window.setTimeout(resolve, ms));
  }

  async function waitForJob(jobId) {
    while (true) {
      const data = await api(`/api/jobs/${jobId}`);
      state.lastJobData = data;
      renderProgress(data);
      updateOverallProgress();
      if (data.status === "done") return data;
      if (data.status === "error" || data.status === "cancelled") {
        const error = new Error(data.error || "");
        error.code = data.error_code || "unknown";
        error.log = data.log || [];
        throw error;
      }
      await delay(1000);
    }
  }

  function applyProgress(percent) {
    const value = Math.max(0, Math.min(100, percent));
    $("#progressPct").textContent = `${Math.round(value)}%`;
    $("#progressBar").style.width = `${value}%`;
    $(".progress").setAttribute("aria-valuenow", String(Math.round(value)));
  }

  function stopProgressAnimation() {
    if (state.progressTimer) {
      window.clearInterval(state.progressTimer);
      state.progressTimer = null;
    }
  }

  function tickProgress() {
    if (!state.running) {
      stopProgressAnimation();
      return;
    }
    const target = state.targetPct;
    let shown = state.displayPct;
    if (shown < target) {
      shown += Math.max((target - shown) * 0.09, 0.05);
    } else {
      const cap = Math.min(99, target + 6);
      if (shown < cap) shown = Math.min(cap, shown + 0.035);
    }
    shown = Math.min(99, Math.max(state.displayPct, shown));
    state.displayPct = shown;
    applyProgress(shown);
  }

  function ensureProgressAnimation() {
    if (state.progressTimer) return;
    state.progressTimer = window.setInterval(tickProgress, 80);
  }

  function updateOverallProgress() {
    const done = state.completed;
    const total = state.total || 1;
    const current =
      state.lastJobData && state.lastJobData.status === "running"
        ? Math.max(0, Math.min(100, Number(state.lastJobData.progress) || 0))
        : 0;
    const percent = Math.max(0, Math.min(100, ((done + current / 100) / total) * 100));
    state.targetPct = Math.min(99.5, percent);
    $("#stageLabel").textContent = `${t("startTest")} ${done}/${total}`;
    $("#elapsed").textContent = formatDuration((Date.now() - state.startedAt) / 1000);
    ensureProgressAnimation();
  }

  function finishTests() {
    stopProgressAnimation();
    state.displayPct = 100;
    state.targetPct = 100;
    applyProgress(100);
    hide($("#progressCard"));
    state.results.sort((a, b) => {
      const subDiff = state.subs.indexOf(a.sub) - state.subs.indexOf(b.sub);
      return subDiff !== 0 ? subDiff : a.pick - b.pick;
    });
    const successes = state.results.filter((item) => item.result);
    if (successes.length) {
      renderResults();
      show($("#resultCard"));
      $("#resultCard").scrollIntoView({ behavior: "smooth", block: "start" });
      return;
    }
    const last = state.results[state.results.length - 1];
    const error = last && last.error ? last.error : { code: "unknown", message: "", log: [] };
    show($("#selectCard"));
    showError(error);
  }

  function resetProgress() {
    stopProgressAnimation();
    state.displayPct = 0;
    state.targetPct = 0;
    $("#stageLabel").textContent = `${t("startTest")} 0/${state.total || 0}`;
    $("#stageMessage").textContent = "";
    $("#stageHint").textContent = "";
    hide($("#stageHint"));
    state.usageWaitStarted = 0;
    applyProgress(0);
    $("#elapsed").textContent = "00:00";
    $("#logBox").textContent = "";
    hide($("#logBox"));
    $("#logToggle").textContent = t("showLog");
  }

  function renderProgress(data) {
    if (!data) return;
    const stageKey = `stage_${data.stage || "queued"}`;
    $("#stageLabel").textContent =
      I18N[state.lang] && I18N[state.lang][stageKey] ? t(stageKey) : t("stage_queued");
    $("#stageMessage").textContent = data.message || "";
    const checkingUsage =
      data.stage === "waiting" && /Checking subscription usage/i.test(data.message || "");
    const hint = $("#stageHint");
    if (checkingUsage) {
      if (!state.usageWaitStarted) state.usageWaitStarted = Date.now();
      if (Date.now() - state.usageWaitStarted >= 6000) {
        hint.textContent = t("usageSlowHint");
        show(hint);
      }
    } else if (state.usageWaitStarted) {
      state.usageWaitStarted = 0;
      hint.textContent = "";
      hide(hint);
    }
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
    if (!state.activeJobs.size) return;
    $("#cancelBtn").disabled = true;
    try {
      await Promise.all(
        [...state.activeJobs].map((jobId) =>
          api(`/api/jobs/${jobId}/cancel`, { method: "POST" }).catch(() => null),
        ),
      );
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
    const hasConfigs = state.subs.some((sub) => sub.configs.length);
    if (hasConfigs) show($("#selectCard"));
    $("#anotherBtn").classList.toggle("hidden", !hasConfigs);
    show($("#errorCard"));
    $("#errorCard").scrollIntoView({ behavior: "smooth", block: "center" });
  }

  function renderResults() {
    const successes = state.results.filter((item) => item.result);
    if (!successes.length) return;
    const count = successes.length;
    const avg =
      successes.reduce((sum, item) => sum + (Number(item.result.ratio) || 0), 0) / count;
    const downloaded = successes.reduce(
      (sum, item) => sum + (Number(item.result.downloaded_bytes) || 0),
      0,
    );
    const deducted = successes.reduce(
      (sum, item) => sum + (Number(item.result.delta_total) || 0),
      0,
    );
    const last = successes[successes.length - 1].result;
    const claimed = state.claimed === undefined ? null : state.claimed;
    const aggregate = {
      ratio: avg,
      downloaded_bytes: downloaded,
      delta_total: deducted,
      claimed_ratio: claimed,
      claimed_matches: null,
      remaining_bytes: last.remaining_bytes,
      config: null,
      core: last.core,
      xray_version: last.xray_version,
      duration_seconds: successes.reduce(
        (sum, item) => sum + (Number(item.result.duration_seconds) || 0),
        0,
      ),
      tested_at: last.tested_at,
      configs_tested: count,
    };
    state.result = aggregate;

    const coef = $("#coef");
    const coefExact = $("#coefExact");
    const verdict = $("#verdict");
    const noRatio = ratioIsNo(avg);
    const unitRatio = ratioIsUnit(avg);
    coef.textContent = ratioBig(avg);
    coef.classList.remove("warn", "bad");
    if (avg > 1.5) coef.classList.add("bad");
    else if (avg > 1.1) coef.classList.add("warn");
    if (noRatio && !unitRatio) {
      coefExact.textContent = t("coefMeasured", { r: formatRatio(avg) });
      show(coefExact);
    } else {
      hide(coefExact);
    }
    verdict.classList.remove("good", "warn", "bad");
    if (noRatio) {
      verdict.textContent = unitRatio ? t("noRatioExact") : t("noRatioClose");
      verdict.classList.add("good");
      show(verdict);
    } else {
      hide(verdict);
    }

    $("#sentence").textContent = t("sentenceAvg", { n: count, r: avg.toFixed(2) });
    $("#statDownloaded").textContent = formatBytes(downloaded);
    $("#statDeducted").textContent = formatBytes(deducted);
    $("#statClaimed").textContent =
      claimed === null || claimed === undefined ? t("notClaimed") : formatRatio(claimed);

    const remainingEl = $("#statRemaining");
    remainingEl.replaceChildren();
    if (last.remaining_bytes === null || last.remaining_bytes === undefined) {
      remainingEl.textContent = "—";
    } else {
      remainingEl.textContent = formatBytes(last.remaining_bytes);
      if (avg > 0) {
        const sub = document.createElement("span");
        sub.className = "stat-sub";
        sub.textContent = t("realEquivalent", { v: formatBytes(last.remaining_bytes / avg) });
        remainingEl.append(sub);
      }
    }

    const meta = $("#resultMeta");
    meta.replaceChildren();
    const rows = [[t("configsToTest"), String(count)]];
    if (aggregate.core) {
      const version = aggregate.xray_version ? ` — ${aggregate.xray_version}` : "";
      rows.push([t("coreUsed"), `${aggregate.core}${version}`]);
    }
    if (aggregate.duration_seconds) {
      rows.push([t("duration"), formatDuration(aggregate.duration_seconds)]);
    }
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

    const list = $("#resultList");
    list.replaceChildren();
    state.results.forEach((item) => {
      const row = document.createElement("div");
      row.className = "result-row";
      const head = document.createElement("div");
      head.className = "result-row-head";
      const name = document.createElement("span");
      name.className = "result-row-name";
      const config = item.sub.configs.find((cfg) => cfg.index === item.pick);
      name.textContent = config
        ? config.remark || config.target || `#${item.pick}`
        : `#${item.pick}`;
      const host = document.createElement("span");
      host.className = "result-row-sub";
      host.textContent = item.sub.host || "";
      head.append(name, host);
      const value = document.createElement("span");
      value.className = "result-row-value";
      if (item.result) {
        const r = Number(item.result.ratio) || 0;
        value.textContent = ratioBig(r);
        const detail = document.createElement("span");
        detail.className = "result-row-detail";
        const bytes = `${formatBytes(item.result.downloaded_bytes)} → ${formatBytes(
          item.result.delta_total,
        )}`;
        detail.textContent =
          ratioIsNo(r) && !ratioIsUnit(r)
            ? `${t("coefMeasured", { r: formatRatio(r) })} · ${bytes}`
            : bytes;
        value.append(detail);
      } else {
        value.textContent = t("resultError");
        value.classList.add("bad");
      }
      row.append(head, value);
      if (item.error) {
        const detail = document.createElement("p");
        detail.className = "hint small";
        detail.textContent = errorMessage(item.error.code);
        row.append(detail);
      }
      list.append(row);
    });
    show($("#resultListWrap"));
  }

  function shareCaption(result) {
    const ratio = Number(result.ratio) || 0;
    const count = Number(result.configs_tested) || 0;
    const sentence =
      count > 1
        ? t("sentenceAvg", { n: count, r: ratio.toFixed(2) })
        : t("sentence", { n: ratio.toFixed(2) });
    const lines = [`${t("shareCaptionTitle")} — ${ratioBig(ratio)}`, sentence];
    if (ratioIsNo(ratio)) {
      lines.push(ratioIsUnit(ratio) ? t("noRatioExact") : t("noRatioClose"));
    }
    lines.push(SITE_URL);
    return lines.join("\n");
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
        button: "#1e40af",
        buttonText: "#ffffff",
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
      button: "#0b3b8c",
      buttonText: "#ffffff",
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
    roundRect(ctx, markX, markY, markSize, markSize, 16);
    ctx.fillStyle = C.button;
    ctx.fill();
    ctx.strokeStyle = C.buttonText;
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

    drawCanvasText(ctx, ratioBig(ratio), rtl ? W - pad : pad, 430, {
      size: 190,
      weight: 900,
      color: ratioColor,
      align: rtl ? "right" : "left",
    });

    let verdictY = 470;
    if (ratioIsNo(ratio)) {
      const unit = ratioIsUnit(ratio);
      if (!unit) {
        drawCanvasText(
          ctx,
          t("coefMeasured", { r: formatRatio(ratio) }),
          rtl ? W - pad : pad,
          486,
          {
            size: 34,
            weight: 700,
            color: C.muted,
            align: rtl ? "right" : "left",
            rtl,
          },
        );
        verdictY = 520;
      }
      drawCanvasText(ctx, unit ? t("noRatioExact") : t("noRatioClose"), rtl ? W - pad : pad, verdictY + 30, {
        size: 30,
        weight: 700,
        color: C.good,
        align: rtl ? "right" : "left",
        rtl,
      });
      verdictY += 64;
    }
    const claimed = result.claimed_ratio;
    if (
      claimed !== null &&
      claimed !== undefined &&
      result.claimed_matches !== null &&
      result.claimed_matches !== undefined
    ) {
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

    const testedCount = Number(result.configs_tested) || 0;
    const sentence =
      testedCount > 1
        ? t("sentenceAvg", { n: testedCount, r: ratio.toFixed(2) })
        : t("sentence", { n: ratio.toFixed(2) });
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

    ctx.fillStyle = C.button;
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
    hide($("#resultListWrap"));
    show($("#selectCard"));
    state.results = [];
    state.completed = 0;
    state.total = 0;
    state.result = null;
    state.error = null;
    state.claimedTouched = false;
    state.claimed = null;
    $("#claimed").value = "";
    renderSubs();
    updateStartState();
    $("#selectCard").scrollIntoView({ behavior: "smooth", block: "start" });
  }

  function retry() {
    hide($("#errorCard"));
    if (selectedCount() > 0) {
      startTest();
      return;
    }
    if (state.subs.some((sub) => sub.configs.length)) {
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
    renderSubs();
    renderClaim();
    updateStartState();
    if (state.result) renderResults();
    if (state.lastJobData && !$("#progressCard").classList.contains("hidden")) {
      renderProgress(state.lastJobData);
      updateOverallProgress();
    }
    if (state.errorCode && !$("#errorCard").classList.contains("hidden")) {
      showError({ code: state.errorCode, message: state.error, log: state.errorLog });
    }
  }

  async function loadVersion() {
    try {
      const res = await fetch("/healthz", { cache: "no-store" });
      if (!res.ok) return;
      const data = await res.json();
      const el = $("#brandVersion");
      if (el && data && data.version) el.textContent = `v${data.version}`;
    } catch (_) {
      /* ignore */
    }
  }

  function init() {
    loadVersion();
    applyI18n();
    applyTheme(document.documentElement.getAttribute("data-theme"), false);
    setSize(5);

    $("#themeBtn").addEventListener("click", toggleTheme);
    $("#langBtn").addEventListener("click", () => {
      setLang(state.lang === "fa" ? "en" : "fa");
    });
    $("#inspectBtn").addEventListener("click", addSubFromInput);
    $("#subUrl").addEventListener("keydown", (event) => {
      if (event.key === "Enter") {
        event.preventDefault();
        addSubFromInput();
      }
    });
    $("#subUrl").addEventListener("paste", handlePaste);
    $("#subUrl").addEventListener("input", (event) => {
      if (event.inputType === "insertFromPaste" || event.inputType === "insertFromDrop") {
        scheduleAutoAddFromPaste();
      }
    });
    $("#sizeRange").addEventListener("input", (event) => setSize(event.target.value));
    $("#testCount").addEventListener("change", (event) => {
      state.testCount = Number(event.target.value) || 1;
      selectConfigsForCount();
    });
    $("#claimed").addEventListener("input", () => {
      state.claimedTouched = true;
    });
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
