import { ref, onMounted, onUnmounted, readonly } from "vue";

export function useMediaQuery(query: string) {
  const matches = ref(false);

  function update() {
    if (typeof window === "undefined" || !window.matchMedia) {
      matches.value = false;
      return;
    }
    matches.value = window.matchMedia(query).matches;
  }

  let mql: MediaQueryList | null = null;

  onMounted(() => {
    update();
    if (typeof window === "undefined" || !window.matchMedia) return;
    mql = window.matchMedia(query);
    mql.addEventListener("change", update);
  });

  onUnmounted(() => {
    if (mql) mql.removeEventListener("change", update);
  });

  return readonly(matches);
}
