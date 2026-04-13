<script setup lang="ts">
import {
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  Title,
  Tooltip,
} from "chart.js";
import { computed } from "vue";
import { Line } from "vue-chartjs";
import { useTelemetryStore } from "@/store/telemetryStore";

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend);

const store = useTelemetryStore();
const points = computed(() => store.snapshot?.recent_points ?? []);
const labels = computed(() => points.value.map((p) => new Date(p.timestamp).toLocaleTimeString()));

const throughputData = computed(() => ({
  labels: labels.value,
  datasets: [
    {
      label: "Throughput",
      data: points.value.map((p) => p.throughput),
      borderColor: "#3498db",
      backgroundColor: "rgba(52,152,219,0.2)",
      tension: 0.2,
    },
  ],
}));

const errorRateData = computed(() => ({
  labels: labels.value,
  datasets: [
    {
      label: "Error Rate",
      data: points.value.map((p) => p.error_rate),
      borderColor: "#e74c3c",
      backgroundColor: "rgba(231,76,60,0.2)",
      tension: 0.2,
    },
  ],
}));

const latencyData = computed(() => ({
  labels: labels.value,
  datasets: [
    {
      label: "Latency p95",
      data: points.value.map((p) => p.latency_p95),
      borderColor: "#8e44ad",
      backgroundColor: "rgba(142,68,173,0.2)",
      tension: 0.2,
    },
  ],
}));

const lagData = computed(() => ({
  labels: labels.value,
  datasets: [
    {
      label: "Consumer Lag",
      data: points.value.map((p) => p.consumer_lag),
      borderColor: "#f39c12",
      backgroundColor: "rgba(243,156,18,0.2)",
      tension: 0.2,
    },
  ],
}));

const options = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: true,
    },
  },
};
</script>

<template>
  <div class="charts">
    <div class="chart-card">
      <h3>Throughput</h3>
      <Line :data="throughputData" :options="options" />
    </div>
    <div class="chart-card">
      <h3>Error Rate</h3>
      <Line :data="errorRateData" :options="options" />
    </div>
    <div class="chart-card">
      <h3>Latency p95</h3>
      <Line :data="latencyData" :options="options" />
    </div>
    <div class="chart-card">
      <h3>Consumer Lag</h3>
      <Line :data="lagData" :options="options" />
    </div>
  </div>
</template>

<style scoped>
.charts {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.chart-card {
  background: #fff;
  border: 1px solid #efefef;
  border-radius: 10px;
  padding: 14px;
  height: 260px;
}

.chart-card h3 {
  margin: 0 0 8px;
  font-size: 14px;
}
</style>

