(function () {
  try {
    var theme = localStorage.getItem("ratio-theme");
    if (theme !== "dark" && theme !== "light") {
      theme = window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
    }
    var root = document.documentElement;
    root.setAttribute("data-theme", theme);
    root.setAttribute("data-mode", theme);
    root.style.colorScheme = theme;
    var lang = localStorage.getItem("ratio-lang");
    if (lang === "en") {
      root.setAttribute("lang", "en");
      root.setAttribute("dir", "ltr");
    }
    var meta = document.querySelector('meta[name="theme-color"]');
    if (meta) meta.setAttribute("content", theme === "dark" ? "#0b0f16" : "#f8fafc");
  } catch (error) {
    /* defaults in the HTML are fine */
  }
})();
