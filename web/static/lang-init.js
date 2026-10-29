(function () {
  var lang = "fa";
  try {
    var saved = localStorage.getItem("ratio-lang");
    if (saved === "fa" || saved === "en") lang = saved;
  } catch (e) {}
  document.documentElement.lang = lang;
  document.documentElement.dir = lang === "fa" ? "rtl" : "ltr";
})();
