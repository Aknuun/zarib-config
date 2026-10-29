(function () {
  var theme = "dark";
  try {
    var saved = localStorage.getItem("ratio-theme");
    if (saved === "dark" || saved === "light") theme = saved;
  } catch (e) {}
  document.documentElement.setAttribute("data-theme", theme);
  var meta = document.querySelector('meta[name="theme-color"]');
  if (meta) meta.setAttribute("content", theme === "dark" ? "#0a0a0a" : "#f6f7f8");
})();
