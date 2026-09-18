import { ref, watch, onMounted } from "vue";

export type Theme = "system" | "light" | "dark";

const STORAGE_KEY = "theme";
const SETTINGS_KEY = "settings";

const theme = ref<Theme>("system");
const isInitialized = ref(false);

function systemPrefersDark(): boolean {
  return (
    typeof window !== "undefined" &&
    window.matchMedia &&
    window.matchMedia("(prefers-color-scheme: dark)").matches
  );
}

function applyThemeClass(value: Theme) {
  if (typeof document === "undefined") return;
  const root = document.documentElement;
  if (value === "dark") {
    root.classList.add("dark");
  } else if (value === "light") {
    root.classList.remove("dark");
  } else {
    root.classList.toggle("dark", systemPrefersDark());
  }
}

function readInitialTheme(): Theme {
  const settingsRaw = localStorage.getItem(SETTINGS_KEY);
  if (settingsRaw) {
    try {
      const s = JSON.parse(settingsRaw);
      if (s.theme === "dark" || s.theme === "light" || s.theme === "system") {
        return s.theme;
      }
    } catch {
      /* ignore */
    }
  }
  const t = localStorage.getItem(STORAGE_KEY);
  if (t === "dark" || t === "light") return t;
  return "system";
}

function init() {
  if (isInitialized.value) return;
  theme.value = readInitialTheme();
  applyThemeClass(theme.value);
  isInitialized.value = true;
}

export function useTheme() {
  onMounted(init);

  watch(theme, (value) => {
    applyThemeClass(value);
    try {
      localStorage.setItem(STORAGE_KEY, value);
    } catch {
      /* ignore */
    }
  });

  function setTheme(value: Theme) {
    if (!isInitialized.value) init();
    theme.value = value;
  }

  function toggle() {
    if (!isInitialized.value) init();
    const current: Theme = theme.value === "system"
      ? systemPrefersDark() ? "light" : "dark"
      : theme.value === "dark" ? "light" : "dark";
    setTheme(current);
  }

  const isDark = ref(false);
  watch(
    theme,
    (value) => {
      isDark.value = value === "dark" || (value === "system" && systemPrefersDark());
    },
    { immediate: true },
  );

  return { theme, setTheme, toggle, isDark };
}
