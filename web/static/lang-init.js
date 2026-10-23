(() => {
  const supported = ["fa", "en"];
  let lang = null;
  try {
    const saved = localStorage.getItem("ratio-lang");
    if (saved === "fa" || saved === "en") lang = saved;
  } catch (_) {
    /* ignore */
  }
  if (!lang) {
    const prefs =
      Array.isArray(navigator.languages) && navigator.languages.length
        ? navigator.languages
        : [navigator.language || ""];
    lang =
      prefs
        .map((value) => String(value).toLowerCase().split("-")[0])
        .find((value) => supported.includes(value)) || "en";
  }
  document.documentElement.lang = lang;
  document.documentElement.dir = lang === "fa" ? "rtl" : "ltr";
})();
