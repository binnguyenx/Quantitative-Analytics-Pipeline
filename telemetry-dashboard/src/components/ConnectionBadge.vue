<script setup lang="ts">
import { computed } from "vue";
import { useTelemetryStore } from "@/store/telemetryStore";

const store = useTelemetryStore();

const color = computed(() => {
  if (store.connection === "connected") return "#2ecc71";
  if (store.connection === "reconnecting") return "#f39c12";
  return "#e74c3c";
});
</script>

<template>
  <div class="badge">
    <span class="dot" :style="{ backgroundColor: color }"></span>
    <span>{{ store.connection }}</span>
  </div>
</template>

<style scoped>
.badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  border: 1px solid #dcdcdc;
  border-radius: 999px;
  padding: 6px 10px;
  font-size: 13px;
  text-transform: uppercase;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
</style>

