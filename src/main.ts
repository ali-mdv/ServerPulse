import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';
import './global.css';
import naive from 'naive-ui';
import VChart from 'vue-echarts';
import * as echarts from 'echarts';
import PrimeVue from 'primevue/config';
import ToastService from 'primevue/toastservice';
import Toast from 'primevue/toast';
import 'primevue/resources/themes/saga-blue/theme.css';
import 'primevue/resources/primevue.min.css';
import 'primeicons/primeicons.css';

const app = createApp(App);
app.use(createPinia());
app.use(router);
app.use(naive);
app.use(PrimeVue);
app.use(ToastService);
app.component('Toast', Toast as any);
app.component('v-chart', VChart as any);
// attach echarts to window for plugins/components that expect it
(window as any).echarts = echarts;
app.mount('#app');
