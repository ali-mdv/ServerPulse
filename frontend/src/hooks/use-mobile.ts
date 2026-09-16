import { ref, onMounted, onUnmounted } from 'vue';

export function useMobile() {
  const isMobile = ref(false);
  function update() { isMobile.value = window.innerWidth < 768; }
  onMounted(() => { update(); window.addEventListener('resize', update); });
  onUnmounted(() => { window.removeEventListener('resize', update); });
  return { isMobile };
}
