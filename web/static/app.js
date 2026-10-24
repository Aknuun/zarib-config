(() => {
  "use strict";

  const I18N = {
    fa: {
      pageTitle: "محاسبه ضریب کانفیگ",
      eyebrow: "شفافیت حجم، بدون حدس",
      heroTitle1: "محاسبه ضریب",
      heroTitle2: "کانفیگ",
      heroSub: "تست دقیق با هم‌ترازی چرخه پنل.",
      privacyNote: "طبق مستندات گیت‌هاب، هیچ کانفیگی نگه‌داشته یا ذخیره نمی‌شود.",
      copyServerScript: "کپی و اجرای مستقیم در سرور خودت",
      copyServerDone: "اسکریپت کپی شد. آن را در ترمینال سرور اجرا کن.",
      copyServerFailed: "کپی نشد؛ متن را دستی انتخاب و کپی کن.",
      chip1: "مگابایت تست",
      chip2: "بدون ذخیره لینک",
      subTitle: "لینک سابسکریپشن",
      statusReady: "آماده",
      subLabel: "لینک سابسکریپشن",
      loadConfigs: "دریافت کانفیگ‌ها",
      pasteSubLink: "جاگذاری لینک ساب",
      pasteHint: "لینک ساب را در کادر بچسبان (Ctrl+V).",
      subLimit: "حداکثر ۵ لینک ساب می‌توانی اضافه کنی.",
      removeSub: "حذف لینک",
      clearLink: "پاک کردن لینک",
      pingAll: "پینگ همه",
      checkSelect: "چک",
      selectMultiple: "انتخاب چند کانفیگ برای تست",
      checkAllTunnels: "چک تانل",
      pingAllDone: "پینگ همه کانفیگ‌ها انجام شد.",
      tunnelAllDone: "وضعیت تانل همه کانفیگ‌ها بررسی شد.",
      subLoading: "در حال خواندن…",
      subConfigs: "{n} کانفیگ",
      configsToTest: "تعداد کانفیگ تست",
      configsShortHint: "کانفیگ‌های هر ساب به‌ترتیب و ساب‌ها هم‌زمان.",
      perConfigResults: "نتیجه هر کانفیگ",
      sentenceAvg: "میانگین ضریب {n} کانفیگ: {r}×",
      err_no_selection: "حداقل یک کانفیگ برای تست انتخاب کن.",
      resultError: "خطا",
      subHint: "لینک‌های قبلی برای استفاده سریع در مرورگر ذخیره می‌شوند.",
      selectTitle: "انتخاب کانفیگ",
      configLabel: "کانفیگ مورد تست",
      tunnelYes: "تانل",
      tunnelNo: "معمولی",
      tunnelUnknown: "نامشخص",
      checkTunnel: "بررسی موقعیت سرور (تانل)",
      checkPing: "پینگ (اتصال TCP)",
      pingOk: "پینگ",
      pingFail: "بدون پاسخ",
      geoFailed: "بررسی تانل ممکن نشد.",
      pingFailed: "پینگ ممکن نشد.",
      tunnelDetected: "سرور در ایران است؛ تانل محسوب می‌شود.",
      tunnelNotDetected: "سرور خارج از ایران است.",
      onlineNow: "آنلاین",
      liveTitle: "آمار زنده",
      liveHeading: "آمار سایت",
      liveClose: "بستن",
      liveHint: "این آمار فقط شمارنده‌های کلی است و هیچ لینک یا کانفیگی ذخیره نمی‌شود.",
      liveToday: "امروز",
      liveMonth: "این ماه",
      liveAll: "کل",
      statOnline: "افراد آنلاین",
      statTests: "تست‌ها",
      statWithRatio: "ضریب‌دار",
      statNormal: "معمولی (۱×)",
      statInvalid: "نامعتبر",
      statDownloaded: "حجم تست‌شده",
      liveTunnelLine: "تانل: {t} · غیرتانل: {n} · نامشخص: {u}",
      advanced: "تنظیمات بیشتر",
      sizeLabel: "حجم تست",
      sizeDialogTitle: "حجم تست",
      sizeRangeHint: "هر چقدر حجم بزرگ‌تر باشد احتمال دقت بیشتر می‌شود.",
      done: "تأیید",
      coreLabel: "هسته Xray",
      coreAuto: "خودکار",
      coreLatest: "جدید",
      coreLegacy: "قدیمی",
      coreSystem: "سیستم",
      coreHint: "اگر کانفیگ با هسته جدید وصل نشد، هسته قدیمی را امتحان کن.",
      bgMeasure: "اندازه‌گیری پس‌زمینه",
      bgMeasureHint: "اگر تیک بزنی، پس‌زمینه را از نتیجه کم می‌کنیم.",
      bgMeasureToast: "تست کمی کندتر پیش می‌رود ولی دقیق‌تر می‌شود.",
      chooseAnother: "کانفیگ دیگری از همین ساب",
      startTest: "شروع تست",
      startConfirmEyebrow: "یادآوری",
      startConfirmHeading: "قبل از شروع تست",
      startConfirmBody: "مطمئن شو هیچ دستگاه یا شخص دیگری از این سابسکریپشن و کانفیگ‌هایش استفاده نمی‌کند. استفاده هم‌زمان باعث می‌شود ضریب با خطا نشان داده شود.",
      startConfirmGo: "مطمئنم، شروع کن",
      privacyHint: "نتیجه فقط در حافظه می‌ماند؛ تاریخچه لینک‌ها در مرورگر خودت ذخیره می‌شود.",
      progressTitle: "در حال تست…",
      cancel: "لغو",
      showLog: "جزئیات",
      hideLog: "بستن",
      stage_queued: "در صف…",
      stage_reading: "خواندن اطلاعات سابسکریپشن…",
      stage_connecting: "اتصال به کانفیگ…",
      stage_baseline: "خواندن مصرف پایه…",
      stage_drift: "هم‌ترازی با چرخه پنل و اندازه‌گیری ترافیک پس‌زمینه…",
      stage_downloading: "دانلود داده تست…",
      stage_waiting: "انتظار برای ثبت مصرف در پنل…",
      stage_usage_pending: "منتظر تأیید برای ادامه…",
      stage_waiting_extended: "در حال انتظار بیشتر برای مصرف…",
      usagePendingHint: "پنل هنوز مصرف را ثبت نکرده است. می‌توانی تا ۵ دقیقه دیگر صبر کنی یا تست را لغو کنی.",
      extendWait: "صبر می‌کنم (۵ دقیقه)",
      extendingWait: "در حال انتظار…",
      stage_done: "تمام شد",
      stage_cancelled: "لغو شد",
      stage_error: "خطا",
      usageSlowHint: "بعضی ساب‌ها دیر نتیجه می‌دهند؛ این ربطی به ماشین‌حساب ضریب ندارد.",
      resultEyebrow: "نتیجه تست",
      resultDisclaimer: "امکان خطا و اشتباه وجود دارد.",
      sentence: "برای هر ۱ مگابایت دانلود، {n} مگابایت از سابسکریپشن شما کم می‌شود.",
      downloaded: "دانلود واقعی",
      deducted: "کسرشده",
      coefMeasured: "اندازه‌گیری واقعی: {r}",
      noRatioClose: "کاملاً بدون ضریب است",
      noRatioExact: "این کانفیگ یا ساب ضریب ندارد",
      nonRoundHint: "ضریب به‌دست‌آمده عدد رند نیست. احتمالاً خودت یا شخص دیگری در حال استفاده از این ساب یا کانفیگ است؛ مطمئن شو کسی استفاده نمی‌کند و دوباره تست بگیر.",
      invalidLow: "تست نامعتبر: مصرف ثبت‌شده کمتر از حجم دانلود است",
      sentenceInvalid: "پنل برای {d} دانلود فقط {u} ثبت کرد؛ این تست معتبر نیست. کانفیگ دیگری را امتحان کن یا دوباره تست بگیر.",
      invalidRow: "تست نامعتبر",
      excludedHint: "{n} کانفیگ به دلیل ثبت‌نشدن مصرف در محاسبه ضریب لحاظ نشد.",
      background: "اصلاح پس‌زمینه",
      backgroundCorrected: "ترافیک پس‌زمینه اندازه‌گیری و {n} از کسرشده کم شد.",
      raw: "خام",
      duration: "مدت تست",
      configUsed: "کانفیگ",
      testAnother: "کانفیگ دیگری از همین ساب",
      shareResult: "اشتراک‌گذاری",
      sharePreparing: "در حال ساخت تصویر…",
      shareFallback: "تصویر نتیجه دانلود شد و متن اشتراک‌گذاری کپی شد. آن را در تلگرام بفرست.",
      shareDownloaded: "تصویر نتیجه دانلود شد. متن اشتراک‌گذاری را دستی کپی کن.",
      shareFailed: "ساخت یا اشتراک‌گذاری تصویر ممکن نشد.",
      reportError: "گزارش مشکل",
      footerReport: "گزارش مشکل",
      reportTitle: "گزارش مشکل",
      reportHeading: "چه مشکلی دیدی؟",
      reportDescriptionLabel: "شرح مشکل",
      reportDescriptionPlaceholder: "مثلاً ضریب اشتباه بود یا تست نامعتبر شد…",
      reportSubLabel: "لینک ساب برای بررسی (حداقل ۱۰۰ مگابایت)",
      reportSubPlaceholder: "https://…",
      reportSubHint: "برای اینکه توسعه‌دهنده بتواند تست را تکرار کند، لینکی بفرست که حداقل ۱۰۰ مگابایت حجم داشته باشد.",
      reportPrivacy: "شرح و لینک از طریق ربات تلگرام برای توسعه‌دهنده ارسال می‌شود.",
      reportSend: "ارسال گزارش",
      reportSending: "در حال ارسال…",
      reportSent: "گزارش ارسال شد. ممنون!",
      reportRequired: "شرح مشکل را بنویس.",
      reportTooLong: "شرح مشکل خیلی طولانی است؛ کوتاه‌ترش کن.",
      reportFailed: "ارسال گزارش ممکن نشد. کمی بعد دوباره تلاش کن.",
      reportDisabled: "ارسال گزارش در حال حاضر فعال نیست.",
      reportInvalidLink: "لینک ساب معتبر نیست.",
      reportRateLimited: "گزارش‌های زیادی فرستادی. کمی بعد دوباره تلاش کن.",
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
      howNote: "این ابزار چرخه به‌روزرسانی پنل را می‌سنجد و اگر تیک بخوری ترافیک پس‌زمینه را از نتیجه کم می‌کند. تست بزرگ‌تر دقیق‌تر است. لینک شما هرگز ذخیره نمی‌شود.",
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
      err_probe_failed: "کانفیگ نتوانست در ۲۰ ثانیه به سرور تست ما وصل شود. پینگ TCP ممکن است کار کند، اما نود ممکن است مقصد تست را مسدود کند یا خیلی کند باشد.",
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
      err_network: "ارتباط با سرور برقرار نشد. اتصال اینترنت یا فیلترشکن را بررسی کن و دوباره تلاش کن.",
      err_aborted: "درخواست قطع شد؛ دوباره تلاش کن.",
      reconnecting: "اتصال قطع شد؛ دوباره تلاش می‌کنم…",
    },
    en: {
      pageTitle: "Calculate config ratio",
      eyebrow: "Transparent data accounting",
      heroTitle1: "Calculate",
      heroTitle2: "config ratio",
      heroSub: "Accurate, panel-synced test.",
      privacyNote: "Per our GitHub docs, no config is retained or stored.",
      copyServerScript: "Copy and run it directly on your server",
      copyServerDone: "Script copied. Paste it into your server terminal.",
      copyServerFailed: "Copy failed; select and copy the text manually.",
      chip1: "MB test",
      chip2: "Link never stored",
      subTitle: "Subscription link",
      statusReady: "Ready",
      subLabel: "Subscription URL",
      loadConfigs: "Load configs",
      pasteSubLink: "Paste sub link",
      pasteHint: "Paste the subscription link in the box (Ctrl+V).",
      subLimit: "You can add up to 5 subscription links.",
      removeSub: "Remove link",
      clearLink: "Clear link",
      pingAll: "Ping all",
      checkSelect: "Check",
      selectMultiple: "Select multiple configs for testing",
      checkAllTunnels: "Check tunnels",
      pingAllDone: "Ping finished for all configs.",
      tunnelAllDone: "Tunnel status checked for all configs.",
      subLoading: "Loading…",
      subConfigs: "{n} configs",
      configsToTest: "Configs to test",
      configsShortHint: "Per-sub configs run in order; subs run in parallel.",
      perConfigResults: "Result per config",
      sentenceAvg: "Average across {n} configs: {r}×",
      err_no_selection: "Select at least one config to test.",
      resultError: "Error",
      subHint: "Previous links are stored in your browser for quick reuse.",
      selectTitle: "Choose a config",
      configLabel: "Config to test",
      tunnelYes: "Tunnel",
      tunnelNo: "Normal",
      tunnelUnknown: "Unknown",
      checkTunnel: "Check server location (tunnel)",
      checkPing: "Ping (TCP connect)",
      pingOk: "Ping",
      pingFail: "No response",
      geoFailed: "Could not check the tunnel status.",
      pingFailed: "Ping failed.",
      tunnelDetected: "The server is in Iran, so it is treated as a tunnel.",
      tunnelNotDetected: "The server is outside Iran.",
      onlineNow: "online",
      liveTitle: "Live stats",
      liveHeading: "Site stats",
      liveClose: "Close",
      liveHint: "These are aggregate counters only; no link or config is stored.",
      liveToday: "Today",
      liveMonth: "This month",
      liveAll: "All time",
      statOnline: "Online now",
      statTests: "Tests",
      statWithRatio: "With ratio",
      statNormal: "Normal (1x)",
      statInvalid: "Invalid",
      statDownloaded: "Tested volume",
      liveTunnelLine: "Tunnel: {t} · Non-tunnel: {n} · Unknown: {u}",
      advanced: "More settings",
      sizeLabel: "Test size",
      sizeDialogTitle: "Test size",
      sizeRangeHint: "The larger the test size, the more accurate the result is likely to be.",
      done: "Done",
      coreLabel: "Xray core",
      coreAuto: "Auto",
      coreLatest: "Latest",
      coreLegacy: "Legacy",
      coreSystem: "System",
      coreHint: "If a config does not connect with the latest core, try the legacy core.",
      bgMeasure: "Background measurement",
      bgMeasureHint: "If enabled, background usage is subtracted from the result.",
      bgMeasureToast: "The test runs a little slower but is more accurate.",
      chooseAnother: "Another config from this sub",
      startTest: "Start test",
      startConfirmEyebrow: "Reminder",
      startConfirmHeading: "Before you start",
      startConfirmBody: "Make sure no other device or person is using this subscription or its configs. Simultaneous use makes the coefficient inaccurate.",
      startConfirmGo: "I'm sure, start",
      privacyHint: "The result stays in memory; previous links are stored in your browser.",
      progressTitle: "Testing…",
      cancel: "Cancel",
      showLog: "Details",
      hideLog: "Close",
      stage_queued: "Queued…",
      stage_reading: "Reading subscription…",
      stage_connecting: "Connecting through the config…",
      stage_baseline: "Reading baseline usage…",
      stage_drift: "Syncing with the panel and measuring background usage…",
      stage_downloading: "Downloading test data…",
      stage_waiting: "Waiting for the panel to record usage…",
      stage_usage_pending: "Waiting for you to continue…",
      stage_waiting_extended: "Waiting longer for usage…",
      usagePendingHint: "The panel has not recorded usage yet. You can wait up to 5 more minutes or cancel the test.",
      extendWait: "Wait 5 more minutes",
      extendingWait: "Waiting…",
      stage_done: "Done",
      stage_cancelled: "Cancelled",
      stage_error: "Error",
      usageSlowHint: "Some subscriptions report usage slowly; this is not caused by the ratio calculator.",
      resultEyebrow: "Test result",
      resultDisclaimer: "Results may contain errors or inaccuracies.",
      sentence: "For every 1 MB downloaded, {n} MB is deducted from your subscription.",
      downloaded: "Downloaded",
      deducted: "Deducted",
      coefMeasured: "Measured: {r}",
      noRatioClose: "Completely without a coefficient",
      noRatioExact: "This config or sub has no coefficient",
      nonRoundHint: "The coefficient is not a round number. You or someone else may be using this subscription or config; make sure it is idle, then test again.",
      invalidLow: "Invalid test: recorded usage is less than downloaded",
      sentenceInvalid: "The panel recorded only {u} for {d} downloaded; this test is invalid. Try another config or retry.",
      invalidRow: "Invalid test",
      excludedHint: "{n} config(s) were excluded from the ratio because their usage was not recorded.",
      background: "Background correction",
      backgroundCorrected: "Background traffic was measured and {n} was subtracted from the deducted amount.",
      raw: "raw",
      duration: "Duration",
      configUsed: "Config",
      testAnother: "Another config from this sub",
      shareResult: "Share",
      sharePreparing: "Creating image…",
      shareFallback: "The result image was downloaded and the share text was copied. Attach it in Telegram.",
      shareDownloaded: "The result image was downloaded. Copy the share text manually.",
      shareFailed: "Could not create or share the result image.",
      reportError: "Report a problem",
      footerReport: "Report a problem",
      reportTitle: "Report a problem",
      reportHeading: "What went wrong?",
      reportDescriptionLabel: "Describe the problem",
      reportDescriptionPlaceholder: "For example, the ratio was wrong or the test was invalid…",
      reportSubLabel: "Subscription link to check (at least 100 MB)",
      reportSubPlaceholder: "https://…",
      reportSubHint: "So the developer can repeat the test, send a link with at least 100 MB of volume.",
      reportPrivacy: "The description and link are sent to the developer through a Telegram bot.",
      reportSend: "Send report",
      reportSending: "Sending…",
      reportSent: "Report sent. Thanks!",
      reportRequired: "Describe the problem first.",
      reportTooLong: "The description is too long; please shorten it.",
      reportFailed: "Could not send the report. Try again later.",
      reportDisabled: "Reporting is not enabled right now.",
      reportInvalidLink: "The subscription link is not valid.",
      reportRateLimited: "You sent too many reports. Try again later.",
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
      howNote: "The tool detects the panel update cycle; if you enable background measurement, it subtracts background traffic. Larger tests are more accurate. Your link is never stored.",
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
      err_probe_failed: "The config could not reach our test server within 20 seconds. TCP ping may still work, but the node may block the test destination or be too slow.",
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
      err_network: "Could not reach the server. Check your internet or VPN and try again.",
      err_aborted: "The request was interrupted; please try again.",
      reconnecting: "Connection lost; retrying…",
    },
  };

  const $ = (sel, root = document) => root.querySelector(sel);
  const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));
  const SITE_URL = "https://zarib.appii.ir/";
  const SUB_HISTORY_KEY = "ratio-sub-history";
  const SUB_HISTORY_LIMIT = 8;
  const LTR_MARK = "\u200E";
  const SELF_HOST_SCRIPT =
    'bash -c "$(curl -fsSL https://github.com/Aknuun/zarib-config/raw/refs/heads/main/run.sh)"';
  const THEME_KEY = "ratio-theme";

  const state = {
    lang: document.documentElement.lang === "en" ? "en" : "fa",
    subs: [],
    size: 25,
    bgMeasure: false,
    sizeCloseTimer: null,
    running: false,
    cancelRequested: false,
    activeJobs: new Set(),
    pendingJobs: new Set(),
    extending: false,
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
    live: null,
    liveId: "",
    liveTimer: null,
    liveRange: "today",
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

  function isNonRoundRatio(value) {
    const n = Number(value);
    if (!Number.isFinite(n) || n <= 0) return false;
    const cents = n * 100;
    return Math.abs(cents - Math.round(cents)) > 0.0001;
  }

  function ratioNumber(value) {
    const n = Number(value) || 0;
    if (ratioIsNo(n)) return "1.00";
    return n.toFixed(isNonRoundRatio(n) ? 3 : 2);
  }

  function formatRatio(value) {
    if (value === null || value === undefined || Number.isNaN(Number(value))) return "—";
    return `${LTR_MARK}×${ratioNumber(value)}`;
  }

  function ratioIsNo(value) {
    const n = Number(value) || 0;
    return n >= 0.95 && n < 1.06;
  }

  function ratioIsInvalid(value) {
    const n = Number(value) || 0;
    return n > 0 && n < 0.95;
  }

  function ratioIsUnit(value) {
    return Math.round((Number(value) || 0) * 100) === 100;
  }

  function ratioBig(value) {
    return ratioIsNo(value) ? `${LTR_MARK}×1.00` : formatRatio(value);
  }

  function resultRatio(item) {
    return item.result ? Number(item.result.ratio) || 0 : null;
  }

  function effectiveRatio(item) {
    return resultRatio(item);
  }

  function aggregateResults(successes) {
    const measured = successes.filter((item) => !ratioIsInvalid(resultRatio(item)));
    const basis = measured.length ? measured : successes;
    let downloaded = 0;
    let deducted = 0;
    let background = 0;
    let weight = 0;
    basis.forEach((item) => {
      const bytes = Number(item.result.downloaded_bytes) || 0;
      const ratio = effectiveRatio(item);
      if (bytes <= 0 || ratio === null) return;
      const measuredRatio = resultRatio(item);
      const bg = Number(item.result.background_bytes_in_test) || 0;
      downloaded += bytes;
      weight += bytes;
      background += bg;
      deducted += ratioIsInvalid(measuredRatio)
        ? ratio * bytes
        : Math.max(0, (Number(item.result.delta_total) || ratio * bytes) - bg);
    });
    return {
      basis,
      excluded: successes.length - basis.length,
      downloaded,
      deducted,
      background,
      ratio: weight > 0 ? deducted / weight : 0,
    };
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

  function randomClientID() {
    try {
      if (window.crypto && typeof window.crypto.randomUUID === "function") {
        return window.crypto.randomUUID();
      }
      const bytes = new Uint8Array(16);
      if (window.crypto && window.crypto.getRandomValues) {
        window.crypto.getRandomValues(bytes);
        return Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
      }
    } catch (_) {
      /* ignore */
    }
    return `anon-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 12)}`;
  }

  function clientID() {
    const key = "ratio-client-id";
    try {
      let id = localStorage.getItem(key);
      if (!id || !/^[A-Za-z0-9_-]{8,64}$/.test(id)) {
        id = randomClientID();
        localStorage.setItem(key, id);
      }
      return id;
    } catch (_) {
      return randomClientID();
    }
  }

  function renderLive() {
    const live = state.live;
    const range = ["today", "month", "all"].includes(state.liveRange)
      ? state.liveRange
      : "today";
    const bucket = (live && live[range === "all" ? "overall" : range === "month" ? "monthly" : "today"]) || {};
    const online = Number(live && live.online) || 0;
    const count = $("#liveCount");
    if (count) count.textContent = live ? String(online) : "—";
    const set = (selector, value) => {
      const el = $(selector);
      if (el) el.textContent = value;
    };
    set("#liveStatOnline", live ? String(online) : "—");
    set("#liveStatTests", live ? String(bucket.tests || 0) : "—");
    set("#liveStatWithRatio", live ? String(bucket.with_ratio || 0) : "—");
    set("#liveStatNormal", live ? String(bucket.normal || 0) : "—");
    set("#liveStatInvalid", live ? String(bucket.invalid || 0) : "—");
    set("#liveStatDownloaded", live ? formatBytes(bucket.downloaded || 0) : "—");
    const line = $("#liveTunnelLine");
    if (line) {
      line.textContent = live
        ? t("liveTunnelLine", {
            t: bucket.tunnel || 0,
            n: bucket.non_tunnel || 0,
            u: bucket.unknown_tunnel || 0,
          })
        : "";
    }
  }

  function setLiveRange(range) {
    state.liveRange = ["today", "month", "all"].includes(range) ? range : "today";
    $$(".live-tab").forEach((tab) => {
      const active = tab.dataset.liveRange === state.liveRange;
      tab.classList.toggle("active", active);
      tab.setAttribute("aria-selected", active ? "true" : "false");
    });
    renderLive();
  }

  async function sendPresence() {
    if (!state.liveId) state.liveId = clientID();
    try {
      const data = await api("/api/presence", {
        method: "POST",
        body: JSON.stringify({ id: state.liveId }),
      });
      state.live = data;
      renderLive();
    } catch (_) {
      /* heartbeats are best-effort */
    }
  }

  function openLiveDialog() {
    const dialog = $("#liveDialog");
    if (!dialog) return;
    sendPresence();
    if (typeof dialog.showModal === "function") {
      if (!dialog.open) dialog.showModal();
    } else {
      dialog.setAttribute("open", "");
    }
  }

  function closeLiveDialog() {
    const dialog = $("#liveDialog");
    if (!dialog) return;
    if (typeof dialog.close === "function") dialog.close();
    else dialog.removeAttribute("open");
  }

  function initLive() {
    state.liveId = clientID();
    sendPresence();
    window.setInterval(() => {
      if (document.visibilityState !== "hidden") sendPresence();
    }, 20000);
    document.addEventListener("visibilitychange", () => {
      if (document.visibilityState === "visible") sendPresence();
    });
    const button = $("#liveBtn");
    if (button) button.addEventListener("click", openLiveDialog);
    const close = $("#liveClose");
    if (close) close.addEventListener("click", closeLiveDialog);
    $$(".live-tab").forEach((tab) => {
      tab.addEventListener("click", () => setLiveRange(tab.dataset.liveRange));
    });
    const dialog = $("#liveDialog");
    if (dialog) {
      dialog.addEventListener("click", (event) => {
        if (event.target === dialog) closeLiveDialog();
      });
    }
  }

  function openReportDialog() {
    const dialog = $("#reportDialog");
    if (!dialog) return;
    const status = $("#reportStatus");
    if (status) {
      status.textContent = "";
      hide(status);
    }
    const subInput = $("#reportSub");
    if (subInput) {
      subInput.value = state.subs.length === 1 ? state.subs[0].url || "" : "";
    }
    if (typeof dialog.showModal === "function") {
      if (!dialog.open) dialog.showModal();
    } else {
      dialog.setAttribute("open", "");
    }
    const description = $("#reportDescription");
    if (description) description.focus();
  }

  function closeReportDialog() {
    const dialog = $("#reportDialog");
    if (!dialog) return;
    if (typeof dialog.close === "function") dialog.close();
    else dialog.removeAttribute("open");
  }

  function openStartDialog() {
    if (state.running) return;
    if (!selectedCount()) {
      showInputError(t("err_no_selection"));
      return;
    }
    const dialog = $("#startDialog");
    if (!dialog) {
      startTest();
      return;
    }
    if (typeof dialog.showModal === "function") {
      if (!dialog.open) dialog.showModal();
    } else {
      dialog.setAttribute("open", "");
    }
    const confirm = $("#startConfirm");
    if (confirm) confirm.focus();
  }

  function closeStartDialog() {
    const dialog = $("#startDialog");
    if (!dialog) return;
    if (typeof dialog.close === "function") dialog.close();
    else dialog.removeAttribute("open");
  }

  function reportErrorText(code) {
    switch (code) {
      case "report_disabled":
        return t("reportDisabled");
      case "invalid_url":
        return t("reportInvalidLink");
      case "rate_limited":
        return t("reportRateLimited");
      case "invalid_input":
        return t("reportRequired");
      case "description_too_long":
        return t("reportTooLong");
      default:
        return t("reportFailed");
    }
  }

  async function submitReport(event) {
    event.preventDefault();
    const descriptionEl = $("#reportDescription");
    const subEl = $("#reportSub");
    const status = $("#reportStatus");
    const description = descriptionEl ? descriptionEl.value.trim() : "";
    const subURL = subEl ? subEl.value.trim() : "";
    if (!description) {
      status.textContent = t("reportRequired");
      show(status);
      descriptionEl.focus();
      return;
    }
    const button = $("#reportSubmit");
    button.disabled = true;
    button.textContent = t("reportSending");
    status.textContent = "";
    hide(status);
    try {
      await api("/api/report", {
        method: "POST",
        body: JSON.stringify({ description, sub_url: subURL, result: state.result }),
      });
      toast(t("reportSent"), 4200);
      const form = $("#reportForm");
      if (form) form.reset();
      closeReportDialog();
    } catch (error) {
      status.textContent = reportErrorText(error.code);
      show(status);
    } finally {
      button.disabled = false;
      button.textContent = t("reportSend");
    }
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
    $("#langBtn").textContent = state.lang === "fa" ? "EN" : "FA";
    applyTheme(document.documentElement.getAttribute("data-theme"), false);
    $$("[data-i18n]").forEach((el) => {
      el.textContent = t(el.dataset.i18n);
    });
    $$("[data-i18n-aria-label]").forEach((el) => {
      const label = t(el.dataset.i18nAriaLabel);
      el.setAttribute("aria-label", label);
      el.setAttribute("title", label);
    });
    $$("[data-i18n-placeholder]").forEach((el) => {
      el.setAttribute("placeholder", t(el.dataset.i18nPlaceholder));
    });
    updateStartLabel();
    updateLogToggle();
    renderLive();
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
    const { retry, timeout = 20000, headers, ...fetchOptions } = options;
    const method = String(fetchOptions.method || "GET").toUpperCase();
    const attempts = Math.max(
      1,
      Number.isFinite(retry) ? retry : method === "GET" ? 3 : 1,
    );
    let lastError = null;
    for (let attempt = 0; attempt < attempts; attempt += 1) {
      const controller = new AbortController();
      const timer = window.setTimeout(() => controller.abort(), timeout);
      try {
        const response = await fetch(path, {
          ...fetchOptions,
          headers: {
            "Content-Type": "application/json",
            ...(headers || {}),
          },
          signal: controller.signal,
        });
        window.clearTimeout(timer);
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
      } catch (error) {
        window.clearTimeout(timer);
        if (error && error.name === "AbortError") {
          error.code = "aborted";
          error.retryable = true;
          throw error;
        }
        if (error && error.code) throw error;
        lastError = error;
        if (lastError && typeof lastError.message === "string") {
          lastError.code = /signal is aborted|Failed to fetch|aborted/i.test(lastError.message)
            ? "network"
            : "unknown";
          lastError.retryable = true;
        }
        if (attempt + 1 < attempts) {
          await delay(600 * (attempt + 1));
          continue;
        }
      }
    }
    const error = new Error("Failed to fetch");
    error.code = "network";
    error.retryable = true;
    if (lastError && lastError.name) error.name = lastError.name;
    throw error;
  }

  function normalizeUrl(raw) {
    let value = raw.trim();
    if (!value) return "";
    if (!/^[a-z][a-z0-9+.-]*:\/\//i.test(value)) value = `https://${value}`;
    return value;
  }

  function updateClearButton() {
    const input = $("#subUrl");
    const button = $("#clearSubUrl");
    if (!input || !button) return;
    button.classList.toggle("hidden", !input.value.trim());
  }

  function loadSubHistory() {
    try {
      const raw = localStorage.getItem(SUB_HISTORY_KEY);
      const items = raw ? JSON.parse(raw) : [];
      if (!Array.isArray(items)) return [];
      return items.filter((item) => typeof item === "string" && item.trim()).slice(0, SUB_HISTORY_LIMIT);
    } catch (_) {
      return [];
    }
  }

  function rememberSubUrl(url) {
    const value = String(url || "").trim();
    if (!value) return;
    try {
      const items = loadSubHistory().filter((item) => item !== value);
      items.unshift(value);
      localStorage.setItem(SUB_HISTORY_KEY, JSON.stringify(items.slice(0, SUB_HISTORY_LIMIT)));
    } catch (_) {
      /* ignore */
    }
  }

  function hideSubHistory() {
    const box = $("#subHistory");
    if (!box) return;
    box.replaceChildren();
    hide(box);
  }

  function renderSubHistory(filter = "") {
    const box = $("#subHistory");
    if (!box) return;
    const query = String(filter || "").trim().toLowerCase();
    const items = loadSubHistory().filter((item) => !query || item.toLowerCase().includes(query));
    if (!items.length) {
      hideSubHistory();
      return;
    }
    box.replaceChildren();
    items.forEach((item) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = "sub-history-item";
      button.setAttribute("role", "option");
      button.textContent = item;
      button.title = item;
      button.addEventListener("click", () => {
        const input = $("#subUrl");
        input.value = item;
        updateClearButton();
        input.focus();
        hideSubHistory();
      });
      box.append(button);
    });
    show(box);
  }

  function showSubHistory() {
    renderSubHistory("");
  }

  function defaultSelection(sub) {
    return sub.configs.slice(0, 1).map((config) => config.index);
  }

  function subHost(url) {
    try {
      return new URL(url).host || url;
    } catch (_) {
      return url;
    }
  }

  function resultIsShowing() {
    return (
      !$("#resultCard").classList.contains("hidden") ||
      !$("#errorCard").classList.contains("hidden")
    );
  }

  function prepareFreshTest() {
    state.subs = [];
    state.results = [];
    state.completed = 0;
    state.total = 0;
    state.result = null;
    state.error = null;
    state.errorCode = null;
    state.errorLog = [];
    stopProgressAnimation();
    hide($("#resultCard"));
    hide($("#errorCard"));
    hide($("#progressCard"));
    hide($("#resultListWrap"));
    hide($("#selectCard"));
  }

  async function addSub(raw, options = {}) {
    if (state.running) return;
    const fresh = options.fresh === true;
    clearInputError();
    const normalized = normalizeUrl(raw);
    if (!normalized) {
      showInputError(errorMessage("invalid_url"));
      return;
    }
    if (!fresh && state.subs.length >= 5) {
      showInputError(t("subLimit"));
      return;
    }
    if (!fresh && state.subs.some((sub) => sub.url === normalized)) {
      $("#subUrl").value = "";
      updateClearButton();
      return;
    }
    if (fresh) prepareFreshTest();
    const sub = {
      id: `sub-${++subSeq}`,
      url: normalized,
      host: subHost(normalized),
      loading: true,
      error: null,
      configs: [],
      count: 0,
      selected: [],
      multi: false,
    };
    state.subs.push(sub);
    $("#subUrl").value = "";
    updateClearButton();
    hideSubHistory();
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
    const fresh = resultIsShowing();
    if (!value) {
      pasteFromClipboard(fresh);
      return;
    }
    const parts = value.split(/\s+/).filter(Boolean);
    if (parts.length > 1) {
      $("#subUrl").value = "";
      if (fresh) prepareFreshTest();
      parts.forEach((part) => addSub(part));
      return;
    }
    addSub(value, { fresh });
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
      updateClearButton();
      if (resultIsShowing()) prepareFreshTest();
      parts.forEach((part) => addSub(part));
      return;
    }
    scheduleAutoAddFromPaste();
  }

  async function pasteFromClipboard(fresh) {
    try {
      if (navigator.clipboard && window.isSecureContext) {
        const text = await navigator.clipboard.readText();
        if (text && text.trim()) {
          $("#subUrl").value = text.trim();
          addSub(text.trim(), { fresh: fresh === true || resultIsShowing() });
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
        retry: 2,
        body: JSON.stringify({ sub_url: sub.url }),
      });
      sub.configs = data.configs || [];
      sub.count = data.count || sub.configs.length;
      sub.selected = defaultSelection(sub);
      rememberSubUrl(sub.url);
      hideSubHistory();
    } catch (error) {
      sub.error = error;
      sub.configs = [];
      sub.count = 0;
      sub.selected = [];
    } finally {
      sub.loading = false;
      renderSubs();
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
    updateStartState();
    if (!state.subs.length) hide($("#selectCard"));
  }

  function configAction(kind, label) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "config-action";
    button.title = label;
    button.setAttribute("aria-label", label);
    if (kind === "geo") {
      button.innerHTML =
        '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" aria-hidden="true"><circle cx="12" cy="12" r="8.5" stroke="currentColor" stroke-width="1.7"/><path d="M3.5 12h17M12 3.5c2.2 2.3 3.3 5.3 3.3 8.5S14.2 18.2 12 20.5C9.8 18.2 8.7 15.2 8.7 12S9.8 5.8 12 3.5Z" stroke="currentColor" stroke-width="1.5"/></svg>';
    } else {
      button.innerHTML =
        '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" aria-hidden="true"><path d="M4 18v-4M9 18V9M14 18v-6M19 18V5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>';
    }
    return button;
  }

  function renderTunnelBadge(badge, config) {
    badge.classList.remove("is-tunnel", "is-normal", "is-unknown");
    if (config.tunnel === true) {
      badge.classList.add("is-tunnel");
      badge.textContent = `✓ ${t("tunnelYes")}`;
      badge.title = config.geoCountry ? `${t("tunnelYes")} · ${config.geoCountry}` : t("tunnelYes");
    } else if (config.tunnel === false) {
      badge.classList.add("is-normal");
      badge.textContent = t("tunnelNo");
      badge.title = config.geoCountry
        ? `${t("tunnelNo")} · ${config.geoCountry}`
        : t("tunnelNo");
    } else {
      badge.classList.add("is-unknown");
      badge.textContent = "—";
      badge.title = t("tunnelUnknown");
    }
  }

  async function checkTunnel(config, badge, button, silent = false) {
    if (button && button.disabled) return;
    if (button) {
      button.disabled = true;
      button.classList.add("busy");
    }
    try {
      const data = await api(`/api/geo?target=${encodeURIComponent(config.target || "")}`);
      config.tunnel = !!data.tunnel;
      config.geoCountry = data.country || "";
      if (badge) renderTunnelBadge(badge, config);
      if (!silent) toast(data.tunnel ? t("tunnelDetected") : t("tunnelNotDetected"));
    } catch (error) {
      if (!silent) toast(error.code === "rate_limited" ? errorMessage("rate_limited") : t("geoFailed"));
    } finally {
      if (button) {
        button.disabled = false;
        button.classList.remove("busy");
      }
    }
  }

  async function runPing(config, output, button, silent = false) {
    if (button && button.disabled) return;
    if (button) {
      button.disabled = true;
      button.classList.add("busy");
    }
    if (output) {
      output.textContent = "…";
      output.classList.remove("ok", "bad");
    }
    try {
      const data = await api(`/api/ping?target=${encodeURIComponent(config.target || "")}`);
      if (data.ok) {
        config.pingMs = Number(data.latency_ms) || 0;
        if (output) {
          output.textContent = `${config.pingMs} ms`;
          output.classList.add("ok");
          output.title = `${t("pingOk")}: ${config.pingMs} ms`;
        }
      } else {
        config.pingMs = null;
        if (output) {
          output.textContent = t("pingFail");
          output.classList.add("bad");
          output.title = t("pingFail");
        }
      }
    } catch (error) {
      config.pingMs = null;
      if (output) {
        output.textContent = t("pingFail");
        output.classList.add("bad");
      }
      if (!silent) toast(error.code === "rate_limited" ? errorMessage("rate_limited") : t("pingFailed"));
    } finally {
      if (button) {
        button.disabled = false;
        button.classList.remove("busy");
      }
    }
  }

  async function runSubAction(sub, mode, button) {
    if (button.disabled) return;
    button.disabled = true;
    const label = button.textContent;
    button.textContent = "…";
    try {
      for (const config of [...sub.configs]) {
        const group = $("#subsContainer").querySelector(`.sub-group[data-sub-id="${sub.id}"]`);
        const item = group
          ? group.querySelector(`.config-item[data-index="${config.index}"]`)
          : null;
        if (!item) continue;
        if (mode === "geo") {
          const badge = item.querySelector(".config-tunnel");
          await checkTunnel(config, badge, null, true);
        } else {
          const output = item.querySelector(".config-ping");
          await runPing(config, output, null, true);
        }
      }
      toast(mode === "geo" ? t("tunnelAllDone") : t("pingAllDone"));
    } finally {
      button.textContent = label;
      button.disabled = state.running;
    }
  }

  function renderSubs() {
    const container = $("#subsContainer");
    const scrolls = new Map();
    container.querySelectorAll(".sub-group").forEach((group) => {
      const list = group.querySelector(".config-list");
      if (list && group.dataset.subId) scrolls.set(group.dataset.subId, list.scrollTop);
    });
    container.replaceChildren();
    state.subs.forEach((sub) => {
      const group = document.createElement("section");
      group.className = "sub-group";
      group.dataset.subId = sub.id;

      const head = document.createElement("div");
      head.className = "sub-head";
      const title = document.createElement("div");
      title.className = "sub-title";
      const nameRow = document.createElement("div");
      nameRow.className = "sub-name-row";
      const host = document.createElement("strong");
      host.textContent = sub.host || sub.url;
      const remove = document.createElement("button");
      remove.type = "button";
      remove.className = "sub-remove";
      remove.textContent = "×";
      remove.title = t("removeSub");
      remove.setAttribute("aria-label", t("removeSub"));
      remove.disabled = state.running;
      remove.addEventListener("click", () => removeSub(sub.id));
      nameRow.append(host, remove);

      const status = document.createElement("span");
      status.className = "sub-status";
      if (sub.loading) status.textContent = t("subLoading");
      else if (sub.error) status.textContent = errorMessage(sub.error.code);
      if (status.textContent) title.append(nameRow, status);
      else title.append(nameRow);

      const actions = document.createElement("div");
      actions.className = "sub-actions";
      const pingAll = document.createElement("button");
      pingAll.type = "button";
      pingAll.className = "ghost-btn sub-action";
      pingAll.textContent = t("pingAll");
      pingAll.disabled = state.running || sub.loading || !!sub.error || !sub.configs.length;
      pingAll.addEventListener("click", () => runSubAction(sub, "ping", pingAll));
      const checkSelect = document.createElement("button");
      checkSelect.type = "button";
      checkSelect.className = "ghost-btn sub-action";
      checkSelect.textContent = sub.multi ? "☑️" : "🔲";
      checkSelect.title = t("selectMultiple");
      checkSelect.setAttribute("aria-label", t("selectMultiple"));
      checkSelect.setAttribute("aria-pressed", sub.multi ? "true" : "false");
      checkSelect.classList.add("icon-only");
      checkSelect.classList.toggle("active", !!sub.multi);
      checkSelect.disabled = state.running || sub.loading || !!sub.error || !sub.configs.length;
      checkSelect.addEventListener("click", () => toggleMulti(sub));
      const tunnelAll = document.createElement("button");
      tunnelAll.type = "button";
      tunnelAll.className = "ghost-btn sub-action";
      tunnelAll.textContent = t("checkAllTunnels");
      tunnelAll.disabled = state.running || sub.loading || !!sub.error || !sub.configs.length;
      tunnelAll.addEventListener("click", () => runSubAction(sub, "geo", tunnelAll));
      actions.append(checkSelect, pingAll, tunnelAll);

      head.append(title, actions);
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
          if (sub.multi) item.classList.add("multi");
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

          const side = document.createElement("span");
          side.className = "config-side";
          const tunnel = document.createElement("span");
          tunnel.className = "config-tunnel";
          renderTunnelBadge(tunnel, config);
          const ping = document.createElement("span");
          ping.className = "config-ping";
          if (config.pingMs !== undefined && config.pingMs !== null) {
            ping.textContent = `${config.pingMs} ms`;
            ping.classList.add("ok");
          } else {
            ping.textContent = "—";
          }
          const geoBtn = configAction("geo", t("checkTunnel"));
          geoBtn.addEventListener("click", (event) => {
            event.preventDefault();
            event.stopPropagation();
            checkTunnel(config, tunnel, geoBtn);
          });
          const pingBtn = configAction("ping", t("checkPing"));
          pingBtn.addEventListener("click", (event) => {
            event.preventDefault();
            event.stopPropagation();
            runPing(config, ping, pingBtn);
          });
          side.append(tunnel, ping, geoBtn, pingBtn);
          item.append(side);

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
    if (!sub.multi) {
      sub.selected = [index];
    } else if (sub.selected.includes(index)) {
      sub.selected = sub.selected.filter((value) => value !== index);
    } else {
      sub.selected = [...sub.selected, index];
    }
    syncSelection(sub);
    updateStartState();
  }

  function toggleMulti(sub) {
    if (state.running) return;
    sub.multi = !sub.multi;
    if (!sub.multi) {
      sub.selected = sub.selected.slice(0, 1);
    }
    if (!sub.selected.length && sub.configs.length) {
      sub.selected = [sub.configs[0].index];
    }
    renderSubs();
    updateStartState();
  }

  function updateLogToggle() {
    $("#logToggle").textContent = state.logOpen ? t("hideLog") : t("showLog");
  }

  function selectedCount() {
    return state.subs.reduce((sum, sub) => sum + sub.selected.length, 0);
  }

  function selectedConfigNames() {
    const names = [];
    state.subs.forEach((sub) => {
      sub.selected.forEach((index) => {
        const config = (sub.configs || []).find((item) => item.index === index);
        names.push(config ? config.remark || config.target || `#${index}` : `#${index}`);
      });
    });
    return names;
  }

  function updateStartLabel() {
    const label = $(".btn-label", $("#startBtn"));
    if (!label) return;
    const names = selectedConfigNames();
    let text = t("startTest");
    if (names.length === 1) text += ` (${names[0]})`;
    else if (names.length > 1) text += ` (${names.length})`;
    label.textContent = text;
    label.title = names.length === 1 ? names[0] : "";
  }

  function updateStartState() {
    updateStartLabel();
    const hasSelection = selectedCount() > 0;
    $("#startBtn").disabled = !hasSelection || state.running;
    $("#inspectBtn").disabled = state.subs.length >= 5 || state.running;
    $("#sizeBtn").disabled = state.running;
    $("#bgToggle").disabled = state.running;
    $("#sizeRange").disabled = state.running;
  }

  function setSize(size) {
    const value = Math.max(5, Math.min(100, Math.round((Number(size) || 25) / 5) * 5));
    state.size = value;
    $("#sizeRange").value = String(value);
    $("#sizeValue").textContent = `${value} MB`;
    const chip = $("#sizeChipValue");
    if (chip) chip.textContent = `${value} MB`;
  }

  function scheduleSizeAutoClose() {
    if (state.sizeCloseTimer) window.clearTimeout(state.sizeCloseTimer);
    state.sizeCloseTimer = window.setTimeout(() => {
      state.sizeCloseTimer = null;
      closeSizeDialog();
    }, 500);
  }

  function openSizeDialog() {
    if (state.running) return;
    const dialog = $("#sizeDialog");
    if (!dialog) return;
    if (state.sizeCloseTimer) {
      window.clearTimeout(state.sizeCloseTimer);
      state.sizeCloseTimer = null;
    }
    const range = $("#sizeRange");
    if (range) range.value = String(state.size);
    if (typeof dialog.showModal === "function") {
      if (!dialog.open) dialog.showModal();
    } else {
      dialog.setAttribute("open", "");
    }
    if (range) range.focus();
  }

  function closeSizeDialog() {
    if (state.sizeCloseTimer) {
      window.clearTimeout(state.sizeCloseTimer);
      state.sizeCloseTimer = null;
    }
    const dialog = $("#sizeDialog");
    if (!dialog) return;
    if (typeof dialog.close === "function") dialog.close();
    else dialog.removeAttribute("open");
  }

  function toggleBgMeasure() {
    if (state.running) return;
    state.bgMeasure = !state.bgMeasure;
    const button = $("#bgToggle");
    if (button) {
      button.classList.toggle("active", state.bgMeasure);
      button.setAttribute("aria-pressed", state.bgMeasure ? "true" : "false");
    }
    if (state.bgMeasure) toast(t("bgMeasureToast"));
  }

  async function startTest() {
    if (state.running) return;
    clearInputError();
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
    state.cancelRequested = false;
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
      await Promise.all(groups.map((sub) => runSubTests(sub)));
    } finally {
      state.running = false;
      state.activeJobs.clear();
      setLoading(button, false);
      updateStartState();
    }
    finishTests();
  }

  async function runSubTests(sub) {
    for (const pick of [...sub.selected]) {
      if (state.cancelRequested) break;
      try {
        const result = await runOne(sub, pick);
        state.results.push({ sub, pick, result, error: null });
      } catch (error) {
        state.results.push({ sub, pick, result: null, error });
      }
      state.completed += 1;
      updateOverallProgress();
    }
  }

  async function runOne(sub, pick) {
    const config = (sub.configs || []).find((item) => item.index === pick);
    const tunnel =
      config && typeof config.tunnel === "boolean" ? config.tunnel : null;
    const data = await api("/api/jobs", {
      method: "POST",
      body: JSON.stringify({
        sub_url: sub.url,
        pick,
        size_mb: state.size,
        wait: 2,
        drift: 0,
        sync: state.bgMeasure,
        core: "auto",
        tunnel,
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
    let networkFailures = 0;
    while (true) {
      let data = null;
      try {
        data = await api(`/api/jobs/${jobId}`, { retry: 0, timeout: 120000 });
        networkFailures = 0;
      } catch (error) {
        const retriable = (
          error.code === "network" ||
          error.code === "aborted" ||
          error.code === "timeout" ||
          error.name === "AbortError"
        );
        if (retriable && !state.cancelRequested) {
          networkFailures += 1;
          const message = $("#stageMessage");
          if (message) message.textContent = t("reconnecting");
          await delay(Math.min(12000, 1000 * networkFailures));
          continue;
        }
        throw error;
      }
      state.lastJobData = data;
      renderProgress(data);
      updateOverallProgress();
      if (data.status === "done") {
        state.pendingJobs.delete(jobId);
        return data;
      }
      if (data.status === "error" || data.status === "cancelled") {
        state.pendingJobs.delete(jobId);
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
    state.pendingJobs.clear();
    state.extending = false;
    hide($("#usagePendingHint"));
    const extendBtn = $("#extendWaitBtn");
    if (extendBtn) {
      hide(extendBtn);
      extendBtn.disabled = false;
      extendBtn.textContent = t("extendWait");
    }
    state.usageWaitStarted = 0;
    applyProgress(0);
    $("#elapsed").textContent = "00:00";
    $("#logBox").textContent = "";
    hide($("#logBox"));
    $("#logToggle").textContent = t("showLog");
  }

  function localizeLogLine(value) {
    const line = String(value || "");
    if (state.lang !== "fa") return line;
    const rules = [
      [/^Reading configs from the subscription \.\.\.$/, "در حال خواندن کانفیگ‌ها از سابسکریپشن ..."],
      [/^Reading subscription info \.\.\.$/, "در حال خواندن اطلاعات سابسکریپشن ..."],
      [/^Current subscription usage: upload (.*) \| download (.*) \| remaining (.*)$/, "مصرف فعلی سابسکریپشن: آپلود $1 | دانلود $2 | باقی‌مانده $3"],
      [/^Current subscription usage: upload (.*) \| download (.*)$/, "مصرف فعلی سابسکریپشن: آپلود $1 | دانلود $2"],
      [/^Using a test-safe copy of the JSON config \(custom routing rules removed\)\.$/, "استفاده از نسخه امن کانفیگ JSON (قوانین مسیریابی سفارشی حذف شد)."],
      [/^Xray \((.+)\) started; probing connectivity through the config \.\.\.$/, "Xray ($1) اجرا شد؛ در حال بررسی اتصال از طریق کانفیگ ..."],
      [/^Connection OK \((.+)\)\. Reading baseline usage \.\.\.$/, "اتصال برقرار شد ($1). خواندن مصرف پایه ..."],
      [/^Connection OK \((.+)\)\.$/, "اتصال برقرار شد ($1)."],
      [/^Measuring background usage across the panel's update cycle \.\.\.$/, "در حال اندازه‌گیری مصرف پس‌زمینه در چرخه به‌روزرسانی پنل ..."],
      [/^No background usage detected\.$/, "مصرف پس‌زمینه‌ای تشخیص داده نشد."],
      [/^Results will be corrected using this background rate\.$/, "نتیجه با این نرخ پس‌زمینه اصلاح می‌شود."],
      [/^Background usage is small \((.+)\) and will be accounted for\.$/, "مصرف پس‌زمینه کم است ($1) و در محاسبه لحاظ می‌شود."],
      [/^Warning: the subscription is consuming data while idle \((.+)\)\.$/, "هشدار: سابسکریپشن در حالت بیکار مصرف دارد ($1)."],
      [/^Primary test target failed; trying fallback (.*) \.\.\.$/, "مقصد اصلی تست ناموفق بود؛ تلاش با مقصد جایگزین $1 ..."],
      [/^Using fallback test target: (.*)$/, "استفاده از مقصد جایگزین تست: $1"],
      [/^Starting test download \((.+)\) \.\.\.$/, "شروع دانلود تست ($1) ..."],
      [/^  received (.*) \| total (.*) of (.*)$/, "  دریافت $1 | مجموع $2 از $3"],
      [/^Download finished: (.*) in (.*)s\.$/, "دانلود تمام شد: $1 در $2 ثانیه."],
      [/^Waiting for the panel to record usage \(up to (.*)s\) \.\.\.$/, "در انتظار ثبت مصرف توسط پنل (تا $1 ثانیه) ..."],
      [/^Checking subscription usage \.\.\.$/, "بررسی مصرف سابسکریپشن ..."],
      [/^Usage reported by the subscription did not change yet\.$/, "مصرف گزارش‌شده توسط سابسکریپشن هنوز تغییر نکرده است."],
      [/^Usage changed by (.*); waiting up to (.*) for the panel to finish recording \.\.\.$/, "مصرف $1 تغییر کرد؛ تا $2 برای تکمیل ثبت پنل صبر می‌کنیم ..."],
      [/^Usage did not change after 120 seconds\. Waiting for user confirmation \.\.\.$/, "مصرف بعد از ۱۲۰ ثانیه تغییر نکرد. در انتظار تأیید کاربر ..."],
      [/^Usage did not change after 120 seconds\.$/, "مصرف بعد از ۱۲۰ ثانیه تغییر نکرد."],
      [/^Continuing to wait for usage \(up to 5 minutes\) \.\.\.$/, "ادامه انتظار برای مصرف (تا ۵ دقیقه) ..."],
      [/^Connection probe failed: (.*)$/, "بررسی اتصال ناموفق بود: $1"],
      [/^Retrying connection(.*)$/, "تلاش دوباره برای اتصال$1"],
      [/^Trying another Xray core \.\.\.$/, "تلاش با هسته دیگر Xray ..."],
      [/^Xray core '(.+)' failed: (.*)$/, "هسته Xray '$1' ناموفق بود: $2"],
      [/^Xray core '(.+)' is unavailable: (.*)$/, "هسته Xray '$1' در دسترس نیست: $2"],
      [/^Xray \((.+)\) not found; downloading from GitHub \((.+)\) \.\.\.$/, "Xray ($1) پیدا نشد؛ دانلود از گیت‌هاب ($2) ..."],
      [/^Warning: only (.*) of (.*) was downloaded\.$/, "هشدار: فقط $1 از $2 دانلود شد."],
      [/^Warning: remaining subscription volume \((.+)\) is less than the test size\.$/, "هشدار: حجم باقی‌مانده سابسکریپشن ($1) از حجم تست کمتر است."],
      [/^Temporary files: (.*)$/, "فایل‌های موقت: $1"],
      [/^JSON result saved to (.*)\.$/, "نتیجه JSON در $1 ذخیره شد."],
      [/^The config did not connect\.$/, "کانفیگ وصل نشد."],
      [/^Could not read usage: (.*)$/, "خواندن مصرف ممکن نشد: $1"],
      [/^  probe failed: (.*)$/, "  بررسی اتصال ناموفق: $1"],
      [/^              COEFFICIENT TEST RESULT$/, "              نتیجه تست ضریب"],
      [/^  REAL COEFFICIENT$/, "  ضریب واقعی"],
      [/^Config: (.*)$/, "کانفیگ: $1"],
      [/^  For every 1 MB downloaded, (.*) is deducted from your subscription\.$/, "  برای هر ۱ مگابایت دانلود، $1 از سابسکریپشن کم می‌شود."],
      [/^  Downloaded\s+(.*)$/, "  دانلود شده $1"],
      [/^  Deducted\s+(.*)$/, "  کسرشده $1"],
      [/^  Claimed\s+not found$/, "  ادعاشده: یافت نشد"],
      [/^  Claimed\s+(.*)$/, "  ادعاشده $1"],
      [/^  Remaining\s+(.*)  \(~(.*) real\)$/, "  باقی‌مانده $1  (~$2 واقعی)"],
      [/^  Remaining\s+(.*)$/, "  باقی‌مانده $1"],
      [/^  Panel recorded\s+(.*)$/, "  ثبت‌شده در پنل $1"],
      [/^  Background\s+(.*)$/, "  پس‌زمینه $1"],
      [/^  background corrected by (.*) \(raw (.*)x\)$/, "  اصلاح پس‌زمینه: $1 (خام $2x)"],
      [/^  Warning: coefficient is below 1; usage may be recorded late\.$/, "  هشدار: ضریب کمتر از ۱ است؛ ممکن است مصرف دیر ثبت شده باشد."],
      [/^Done$/, "تمام شد"],
      [/^Cancelled\.$/, "لغو شد."],
    ];
    for (const [pattern, replacement] of rules) {
      if (pattern.test(line)) return line.replace(pattern, replacement);
    }
    return line;
  }

  function renderProgress(data) {
    if (!data) return;
    const pending = data.stage === "usage_pending";
    if (data.id) {
      if (pending) state.pendingJobs.add(data.id);
      else state.pendingJobs.delete(data.id);
    }
    const pendingHint = $("#usagePendingHint");
    if (pendingHint) {
      if (pending) show(pendingHint);
      else hide(pendingHint);
    }
    const extendBtn = $("#extendWaitBtn");
    if (extendBtn) {
      if (pending) {
        show(extendBtn);
        extendBtn.disabled = state.extending;
        extendBtn.textContent = t(state.extending ? "extendingWait" : "extendWait");
      } else {
        hide(extendBtn);
        extendBtn.disabled = false;
        extendBtn.textContent = t("extendWait");
        state.extending = false;
      }
    }
    const stageKey = `stage_${data.stage || "queued"}`;
    $("#stageLabel").textContent =
      I18N[state.lang] && I18N[state.lang][stageKey] ? t(stageKey) : t("stage_queued");
    $("#stageMessage").textContent = localizeLogLine(data.message || "");
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
      const log = (data.log || []).map(localizeLogLine).join("\n");
      const box = $("#logBox");
      box.textContent = log;
      box.scrollTop = box.scrollHeight;
    }
  }

  async function extendWaiting() {
    const ids = [...state.pendingJobs];
    if (!ids.length || state.extending) return;
    state.extending = true;
    const button = $("#extendWaitBtn");
    if (button) {
      button.disabled = true;
      button.textContent = t("extendingWait");
    }
    try {
      await Promise.all(
        ids.map((jobId) =>
          api(`/api/jobs/${jobId}/extend`, { method: "POST" }).catch(() => null),
        ),
      );
    } finally {
      state.extending = false;
    }
  }

  async function cancelTest() {
    state.cancelRequested = true;
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
    const detail = [localizeLogLine(state.error), ...state.errorLog.slice(-20).map(localizeLogLine)].filter(Boolean).join("\n");
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
    const agg = aggregateResults(successes);
    const avg = agg.ratio;
    const downloaded = agg.downloaded;
    const deducted = agg.deducted;
    const usableCount = agg.basis.length || count;
    const nonRound = successes.some((item) =>
      isNonRoundRatio(item.result.ratio_raw ?? item.result.ratio),
    );
    const last = successes[successes.length - 1].result;
    const aggregate = {
      ratio: avg,
      downloaded_bytes: downloaded,
      delta_total: deducted,
      config: null,
      core: last.core,
      xray_version: last.xray_version,
      duration_seconds: successes.reduce(
        (sum, item) => sum + (Number(item.result.duration_seconds) || 0),
        0,
      ),
      tested_at: last.tested_at,
      configs_tested: usableCount,
    };
    state.result = aggregate;

    const coef = $("#coef");
    const coefExact = $("#coefExact");
    const verdict = $("#verdict");
    const noRatio = ratioIsNo(avg);
    const unitRatio = ratioIsUnit(avg);
    const invalid = ratioIsInvalid(avg);
    const glyph = $("#resultGlyph");
    if (glyph) {
      glyph.textContent = noRatio ? "✓" : "×";
      glyph.classList.toggle("is-check", noRatio);
      glyph.classList.toggle("is-multiply", !noRatio);
    }
    coef.textContent = ratioBig(avg);
    coef.classList.remove("warn", "bad");
    if (invalid || avg > 1.5) coef.classList.add("bad");
    else if (avg > 1.1) coef.classList.add("warn");
    if (invalid || (noRatio && !unitRatio)) {
      coefExact.textContent = t("coefMeasured", { r: formatRatio(avg) });
      show(coefExact);
    } else {
      hide(coefExact);
    }
    const inlineShare = $("#shareInlineBtn");
    if (inlineShare) {
      if (noRatio) show(inlineShare);
      else hide(inlineShare);
    }
    verdict.classList.remove("good", "warn", "bad");
    if (invalid) {
      verdict.textContent = t("invalidLow");
      verdict.classList.add("bad");
      show(verdict);
    } else if (noRatio) {
      verdict.textContent = unitRatio ? t("noRatioExact") : t("noRatioClose");
      verdict.classList.add("good");
      show(verdict);
    } else {
      hide(verdict);
    }

    $("#sentence").textContent = invalid
      ? t("sentenceInvalid", { d: formatBytes(downloaded), u: formatBytes(deducted) })
      : count > 1
        ? t("sentenceAvg", { n: usableCount, r: ratioNumber(avg) })
        : t("sentence", { n: ratioNumber(avg) });
    $("#statDownloaded").textContent = formatBytes(downloaded);
    $("#statDeducted").textContent = formatBytes(deducted);

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
        const rowInvalid = ratioIsInvalid(r);
        value.textContent = ratioBig(r);
        if (rowInvalid) value.classList.add("bad");
        const detail = document.createElement("span");
        detail.className = "result-row-detail";
        const deducted = Math.max(
          0,
          (Number(item.result.delta_total) || 0) -
            (Number(item.result.background_bytes_in_test) || 0),
        );
        const bytes = `${formatBytes(item.result.downloaded_bytes)} → ${formatBytes(deducted)}`;
        if (rowInvalid) {
          detail.textContent = `${t("invalidRow")} · ${bytes}`;
        } else if (ratioIsNo(r) && !ratioIsUnit(r)) {
          detail.textContent = `${t("coefMeasured", { r: formatRatio(r) })} · ${bytes}`;
        } else {
          detail.textContent = bytes;
        }
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
    const note = $("#resultNote");
    if (note) {
      const notes = [];
      if (agg.excluded > 0) notes.push(t("excludedHint", { n: agg.excluded }));
      if (agg.background > 0) {
        notes.push(t("backgroundCorrected", { n: formatBytes(agg.background) }));
      }
      if (nonRound) notes.push(t("nonRoundHint"));
      if (notes.length) {
        note.textContent = notes.join(" ");
        show(note);
      } else {
        note.textContent = "";
        hide(note);
      }
    }
    show($("#resultListWrap"));
  }

  function shareCaption(result) {
    const ratio = Number(result.ratio) || 0;
    const count = Number(result.configs_tested) || 0;
    const sentence =
      count > 1
        ? t("sentenceAvg", { n: count, r: ratioNumber(ratio) })
        : t("sentence", { n: ratioNumber(ratio) });
    const lines = [`${t("shareCaptionTitle")} — ${ratioBig(ratio)}`, sentence];
    if (ratioIsInvalid(ratio)) {
      lines.push(t("invalidLow"));
    } else if (ratioIsNo(ratio)) {
      lines.push(ratioIsUnit(ratio) ? t("noRatioExact") : t("noRatioClose"));
    }
    lines.push(t("resultDisclaimer"));
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
    if (ratioIsInvalid(ratio) || ratioIsNo(ratio)) {
      const invalid = ratioIsInvalid(ratio);
      const unit = ratioIsUnit(ratio);
      if (invalid || !unit) {
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
      drawCanvasText(
        ctx,
        invalid ? t("invalidLow") : unit ? t("noRatioExact") : t("noRatioClose"),
        rtl ? W - pad : pad,
        verdictY + 30,
        {
          size: invalid ? 26 : 30,
          weight: 700,
          color: invalid ? C.danger : C.good,
          align: rtl ? "right" : "left",
          rtl,
        },
      );
      verdictY += 64;
    }
    const testedCount = Number(result.configs_tested) || 0;
    const sentence =
      testedCount > 1
        ? t("sentenceAvg", { n: testedCount, r: ratioNumber(ratio) })
        : t("sentence", { n: ratioNumber(ratio) });
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
    const stats = [
      { label: t("downloaded"), value: formatBytes(result.downloaded_bytes) },
      { label: t("deducted"), value: formatBytes(result.delta_total) },
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

    const statRows = Math.ceil(stats.length / 2);
    const metaY = statsTop + statRows * statH + (statRows - 1) * statGap + 44;
    const metaLines = [];
    if (result.config) metaLines.push(`${t("configUsed")}: ${result.config}`);
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

  async function shareResult(buttonOverride = null) {
    if (!state.result) return;
    const button = buttonOverride || $("#shareBtn");
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
    setSize(25);
    updateClearButton();
    initLive();

    $("#themeBtn").addEventListener("click", toggleTheme);
    $("#langBtn").addEventListener("click", () => {
      setLang(state.lang === "fa" ? "en" : "fa");
    });
    $("#inspectBtn").addEventListener("click", addSubFromInput);
    $("#subUrl").addEventListener("keydown", (event) => {
      if (event.key === "Enter") {
        event.preventDefault();
        addSubFromInput();
      } else if (event.key === "Escape") {
        hideSubHistory();
      }
    });
    $("#subUrl").addEventListener("focus", showSubHistory);
    $("#subUrl").addEventListener("click", showSubHistory);
    const clearSubUrl = $("#clearSubUrl");
    if (clearSubUrl) {
      clearSubUrl.addEventListener("click", () => {
        const input = $("#subUrl");
        input.value = "";
        updateClearButton();
        input.focus();
        hideSubHistory();
      });
    }
    $("#subUrl").addEventListener("paste", handlePaste);
    $("#subUrl").addEventListener("input", (event) => {
      updateClearButton();
      renderSubHistory(event.target.value);
      if (event.inputType === "insertFromPaste" || event.inputType === "insertFromDrop") {
        scheduleAutoAddFromPaste();
      }
    });
    document.addEventListener("click", (event) => {
      if (!event.target.closest(".input-wrap")) hideSubHistory();
    });
    const sizeRange = $("#sizeRange");
    sizeRange.addEventListener("input", (event) => {
      setSize(event.target.value);
      scheduleSizeAutoClose();
    });
    sizeRange.addEventListener("change", (event) => {
      setSize(event.target.value);
      closeSizeDialog();
    });
    sizeRange.addEventListener("wheel", (event) => {
      if (state.running) return;
      event.preventDefault();
      setSize(state.size + (event.deltaY > 0 ? -5 : 5));
      scheduleSizeAutoClose();
    }, { passive: false });
    $("#sizeBtn").addEventListener("click", openSizeDialog);
    const sizeClose = $("#sizeClose");
    if (sizeClose) sizeClose.addEventListener("click", closeSizeDialog);
    const sizeDialog = $("#sizeDialog");
    if (sizeDialog) {
      sizeDialog.addEventListener("click", (event) => {
        if (event.target === sizeDialog) closeSizeDialog();
      });
    }
    const bgToggle = $("#bgToggle");
    if (bgToggle) bgToggle.addEventListener("click", toggleBgMeasure);
    $("#startBtn").addEventListener("click", openStartDialog);
    const startConfirm = $("#startConfirm");
    if (startConfirm) {
      startConfirm.addEventListener("click", () => {
        closeStartDialog();
        startTest();
      });
    }
    const startCancel = $("#startCancel");
    if (startCancel) startCancel.addEventListener("click", closeStartDialog);
    const startClose = $("#startClose");
    if (startClose) startClose.addEventListener("click", closeStartDialog);
    const startDialog = $("#startDialog");
    if (startDialog) {
      startDialog.addEventListener("click", (event) => {
        if (event.target === startDialog) closeStartDialog();
      });
    }
    $("#cancelBtn").addEventListener("click", cancelTest);
    const extendBtn = $("#extendWaitBtn");
    if (extendBtn) extendBtn.addEventListener("click", extendWaiting);
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
    $("#shareBtn").addEventListener("click", () => shareResult());
    const shareInlineBtn = $("#shareInlineBtn");
    if (shareInlineBtn) shareInlineBtn.addEventListener("click", () => shareResult(shareInlineBtn));
    $("#reportBtn").addEventListener("click", openReportDialog);
    const footerReport = $("#footerReportBtn");
    if (footerReport) footerReport.addEventListener("click", openReportDialog);
    $("#reportClose").addEventListener("click", closeReportDialog);
    $("#reportCancel").addEventListener("click", closeReportDialog);
    $("#reportForm").addEventListener("submit", submitReport);
    const reportDialog = $("#reportDialog");
    if (reportDialog) {
      reportDialog.addEventListener("click", (event) => {
        if (event.target === reportDialog) closeReportDialog();
      });
    }
    const copyScriptBtn = $("#copyScriptBtn");
    if (copyScriptBtn) {
      copyScriptBtn.addEventListener("click", async () => {
        const copied = await copyText(SELF_HOST_SCRIPT);
        toast(copied ? t("copyServerDone") : t("copyServerFailed"), copied ? 4200 : 6500);
      });
    }
    $("#retryBtn").addEventListener("click", retry);
    $("#anotherBtn").addEventListener("click", testAnother);
  }

  init();
})();
