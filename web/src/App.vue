<script setup>
import { computed, onMounted, ref } from "vue";
import NewIssueModal from "./components/NewIssueModal.vue";

const issues = ref([]);
const error = ref(null);
const showNewIssueModal = ref(false);

const expandedId = ref(null);
const subtasksById = ref({});
const subtasksLoadingId = ref(null);
const subtasksErrorById = ref({});

const columns = [
  { status: "open", label: "Open", accent: "border-t-blue-400", badge: "bg-blue-100 text-blue-800" },
  { status: "in-progress", label: "In Progress", accent: "border-t-amber-400", badge: "bg-amber-100 text-amber-800" },
  { status: "done", label: "Done", accent: "border-t-green-400", badge: "bg-green-100 text-green-800" },
];

const issuesByStatus = computed(() => {
  const groups = Object.fromEntries(columns.map((c) => [c.status, []]));
  for (const issue of issues.value) {
    (groups[issue.status] ?? (groups[issue.status] = [])).push(issue);
  }
  return groups;
});

onMounted(async () => {
  try {
    const res = await fetch("/api/issues");
    if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
    issues.value = await res.json();
  } catch (e) {
    error.value = e.message;
  }
});

function onIssueCreated(issue) {
  issues.value.push(issue);
  showNewIssueModal.value = false;
}

function formatDate(value) {
  return new Date(value).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
}

async function toggleExpanded(id) {
  if (expandedId.value === id) {
    expandedId.value = null;
    return;
  }

  expandedId.value = id;
  if (subtasksById.value[id]) return;

  subtasksLoadingId.value = id;
  try {
    const res = await fetch(`/api/issues/${id}`);
    if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
    const detail = await res.json();
    subtasksById.value[id] = detail.subtasks ?? [];
  } catch (e) {
    subtasksErrorById.value[id] = e.message;
  } finally {
    subtasksLoadingId.value = null;
  }
}
</script>

<template>
  <main class="mx-auto max-w-6xl p-6">
    <h1 class="mb-6 text-2xl font-semibold text-gray-900">Issues</h1>

    <p v-if="error" class="text-red-600">Failed to load issues: {{ error }}</p>

    <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
      <section
        v-for="col in columns"
        :key="col.status"
        class="flex flex-col rounded-lg border border-gray-200 border-t-4 bg-gray-50"
        :class="col.accent"
      >
        <header class="flex items-center justify-between px-4 py-3">
          <h2 class="text-sm font-semibold tracking-wide text-gray-700 uppercase">
            {{ col.label }}
          </h2>
          <div class="flex items-center gap-2">
            <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="col.badge">
              {{ issuesByStatus[col.status].length }}
            </span>
            <button
              v-if="col.status === 'open'"
              type="button"
              class="rounded-full bg-blue-600 px-2 py-0.5 text-xs font-medium text-white hover:bg-blue-700"
              @click="showNewIssueModal = true"
            >
              + Add
            </button>
          </div>
        </header>

        <div class="flex flex-col gap-3 px-3 pb-3">
          <article
            v-for="issue in issuesByStatus[col.status]"
            :key="issue.id"
            class="cursor-pointer rounded-lg border border-gray-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md"
            @click="toggleExpanded(issue.id)"
          >
            <div class="mb-2 text-xs font-mono text-gray-400">#{{ issue.id }}</div>
            <h3 class="font-medium text-gray-900">{{ issue.title }}</h3>
            <p v-if="issue.description" class="mt-2 line-clamp-3 text-sm text-gray-500">
              {{ issue.description }}
            </p>
            <div class="mt-2 text-xs text-gray-400">
              <span>Created {{ formatDate(issue.created_at) }}</span>
              <span v-if="issue.completed_at"> · Completed {{ formatDate(issue.completed_at) }}</span>
            </div>

            <div v-if="expandedId === issue.id" class="mt-3 border-t border-gray-100 pt-3">
              <p v-if="subtasksLoadingId === issue.id" class="text-xs text-gray-400 italic">
                Loading subtasks…
              </p>
              <p v-else-if="subtasksErrorById[issue.id]" class="text-xs text-red-500">
                Failed to load subtasks: {{ subtasksErrorById[issue.id] }}
              </p>
              <ul v-else-if="subtasksById[issue.id]?.length" class="flex flex-col gap-1">
                <li
                  v-for="subtask in subtasksById[issue.id]"
                  :key="subtask.text"
                  class="flex items-center gap-2 text-sm text-gray-600"
                >
                  <span>{{ subtask.done ? "☑" : "☐" }}</span>
                  <span :class="{ 'text-gray-400 line-through': subtask.done }">{{ subtask.text }}</span>
                </li>
              </ul>
              <p v-else class="text-xs text-gray-400 italic">No subtasks</p>
            </div>
          </article>

          <p v-if="issuesByStatus[col.status].length === 0" class="px-1 py-2 text-sm text-gray-400 italic">
            No issues
          </p>
        </div>
      </section>
    </div>

    <NewIssueModal
      v-if="showNewIssueModal"
      @close="showNewIssueModal = false"
      @created="onIssueCreated"
    />
  </main>
</template>
