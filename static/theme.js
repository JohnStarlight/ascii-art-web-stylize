document.addEventListener("DOMContentLoaded", () => {
  const themes = {
    ppetraki: "theme-ai",
    balevizos: "theme-pink",
    ivogiagke: "theme-keys"
  };

  const textInput = document.getElementById("text");
  const themeSelect = document.getElementById("theme-select");

  if (!textInput || !themeSelect) return;

  let currentTheme = localStorage.getItem("ascii-theme") || "theme-guest";
  document.body.classList.remove("theme-guest", "theme-ai", "theme-pink", "theme-keys");
  document.body.classList.add(currentTheme);
  themeSelect.value = currentTheme;

  function applyTheme(theme) {
    if (theme === currentTheme) return;
    document.body.classList.remove(currentTheme);
    document.body.classList.add(theme);
    currentTheme = theme;
    localStorage.setItem("ascii-theme", theme);
    themeSelect.value = theme;
  }

  themeSelect.addEventListener("change", () => {
    applyTheme(themeSelect.value);
  });

  textInput.addEventListener("input", () => {
    const value = textInput.value.trim().toLowerCase();
    if (themes[value]) {
      applyTheme(themes[value]);
    }
  });
});