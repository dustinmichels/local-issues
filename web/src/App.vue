<script setup>
import { onMounted, ref } from "vue";

const issues = ref([]);
const error = ref(null);

const statusStyles = {
  open: "bg-blue-100 text-blue-800",
  "in-progress": "bg-amber-100 text-amber-800",
  done: "bg-green-100 text-green-800",
};

function statusClass(status) {
  return statusStyles[status] ?? "bg-gray-100 text-gray-800";
}

onMounted(async () => {
  try {
    const res = await fetch("/api/issues");
    if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
    issues.value = await res.json();
  } catch (e) {
    error.value = e.message;
  }
});
</script>

<template>
  <main class="mx-auto max-w-5xl p-6">
    <h1 class="mb-6 text-2xl font-semibold text-gray-900">Issues</h1>

    <p v-if="error" class="text-red-600">Failed to load issues: {{ error }}</p>

    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="issue in issues"
        :key="issue.id"
        class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm"
      >
        <div class="mb-2 text-xs font-mono text-gray-400">#{{ issue.id }}</div>
        <h2 class="mb-3 font-medium text-gray-900">{{ issue.title }}</h2>
        <span
          class="inline-block rounded-full px-2 py-1 text-xs font-medium"
          :class="statusClass(issue.status)"
        >
          {{ issue.status }}
        </span>
      </div>
    </div>
  </main>
</template>
