<template>
  <div class="space-y-6">
    <div class="text-sm text-muted">Window: {{ Math.round(width) }} × {{ Math.round(height) }}</div>
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <Card>
        <div>
          <div class="text-sm text-muted">Total Servers</div>
          <div class="text-2xl font-bold">12</div>
        </div>
      </Card>
      <Card>
        <div>
          <div class="text-sm text-muted">Online</div>
          <div class="text-2xl font-bold text-green-600">10</div>
        </div>
      </Card>
      <Card>
        <div>
          <div class="text-sm text-muted">Down</div>
          <div class="text-2xl font-bold text-red-600">2</div>
        </div>
      </Card>
      <Card>
        <div>
          <div class="text-sm text-muted">High Load</div>
          <div class="text-2xl font-bold text-yellow-600">3</div>
        </div>
      </Card>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <Card class="col-span-2">
        <h3 class="font-semibold mb-3">CPU Usage (Total)</h3>
        <v-chart :option="cpuOption" style="height:260px; width:100%" />
      </Card>

      <Card>
        <h3 class="font-semibold mb-3">Memory Usage</h3>
        <v-chart :option="memOption" style="height:220px; width:100%" />
      </Card>
    </div>

    <div>
      <h3 class="font-semibold mb-3">Servers with High Usage</h3>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card v-for="s in highServers" :key="s.id">
          <div class="font-bold">{{ s.name }}</div>
          <div class="text-sm text-muted">CPU: {{ s.cpu }}% • Mem: {{ s.mem }}%</div>
        </Card>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import Card from '@/components/Card.vue';
import { reactive } from 'vue';
import { useWindowSize } from '@vueuse/core';

const { width, height } = useWindowSize();

const cpuData = reactive({
  labels: Array.from({ length: 12 }, (_, i) => `${i}:00`),
  datasets: [
    { label: 'CPU %', data: Array.from({ length: 12 }, () => Math.round(Math.random() * 100)), borderColor: '#2563EB', backgroundColor: 'rgba(37,99,235,0.1)' }
  ]
});

import { computed } from 'vue';
const cpuOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  xAxis: { type: 'category', data: cpuData.labels },
  yAxis: { type: 'value' },
  series: [
    {
      name: cpuData.datasets[0].label,
      type: 'line',
      data: cpuData.datasets[0].data,
      smooth: true,
      lineStyle: { color: cpuData.datasets[0].borderColor },
      areaStyle: { color: cpuData.datasets[0].backgroundColor }
    }
  ]
}));

const memoryData = reactive({
  labels: ['Used', 'Free'],
  datasets: [{ data: [65, 35], backgroundColor: ['#60A5FA', '#E5E7EB'] }]
});

const lineOptions = { responsive: true, maintainAspectRatio: false };

// vue-echarts memory pie option
const memOption = reactive({
  tooltip: { trigger: 'item' },
  legend: { bottom: 0 },
  series: [
    {
      name: 'Memory',
      type: 'pie',
      radius: '60%',
      data: [
        { value: 65, name: 'Used' },
        { value: 35, name: 'Free' }
      ],
      emphasis: { itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0,0,0,0.5)' } }
    }
  ]
});

const highServers = [
  { id: '1', name: 'web-01', cpu: 92, mem: 78 },
  { id: '2', name: 'db-01', cpu: 88, mem: 85 },
  { id: '3', name: 'cache-01', cpu: 79, mem: 61 },
];
</script>
