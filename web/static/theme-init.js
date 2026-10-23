(function () {
  try {
    var theme = localStorage.getItem("ratio-theme");
    if (theme === "dark" || theme === "light") {
      document.documentElement.setAttribute("data-theme", theme);
      var meta = document.querySelector('meta[name="theme-color"]');
      if (meta) meta.setAttribute("content", theme === "dark" ? "#050810" : "#f5f8fc");
    }
  } catch (e) {
    /* ignore */
  }
})();
