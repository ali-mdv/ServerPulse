import { ref, onMounted, onUnmounted } from "vue";
import { readChartTheme, type ChartTheme } from "@/lib/chart-theme";

export function useChartTheme() {
  const theme = ref<ChartTheme>(readChartTheme());

  function refresh() {
    theme.value = readChartTheme();
  }

  let observer: MutationObserver | null = null;

  onMounted(() => {
    refresh();
    if (typeof document !== "undefined") {
      observer = new MutationObserver(refresh);
      observer.observe(document.documentElement, {
        attributes: true,
        attributeFilter: ["class"],
      });
    }
  });

  onUnmounted(() => {
    if (observer) observer.disconnect();
  });

  return theme;
}
