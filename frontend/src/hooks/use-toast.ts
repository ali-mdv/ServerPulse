import { ref } from 'vue';

const toasts = ref<string[]>([]);
export function useToast() { return { toasts, push: (t:string)=> toasts.value.push(t), remove: (i:number)=> toasts.value.splice(i,1) } }
