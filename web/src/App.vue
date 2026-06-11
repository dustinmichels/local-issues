<script setup>
import { computed, onMounted, ref } from "vue";
import NewIssueModal from "./components/NewIssueModal.vue";
import EditIssueModal from "./components/EditIssueModal.vue";

const issues = ref([]);
const error = ref(null);
const actionError = ref(null);
const showNewIssueModal = ref(false);
const editingIssueId = ref(null);
const frozen = ref(false);

const expandedId = ref(null);
const subtasksById = ref({});
const subtasksLoadingId = ref(null);
const subtasksErrorById = ref({});

const draggedId = ref(null);
const dragOverStatus = ref(null);

const columns = [
  { status: "open", label: "Open", accent: "border-t-blue-400", badge: "bg-blue-100 text-blue-800" },
  { status: "in-progress", label: "In Progress", accent: "border-t-amber-400", badge: "bg-amber-100 text-amber-800" },
  { status: "done", label: "Done", accent: "border-t-green-400", badge: "bg-green-100 text-green-800" },
];

const filterOptions = [
  { key: "created-today", label: "Created: Today" },
  { key: "completed-today", label: "Completed: Today" },
  { key: "created-week", label: "Created: This Week" },
  { key: "completed-week", label: "Completed: This Week" },
];

const activeFilter = ref(null);

function isToday(value) {
  const d = new Date(value);
  const now = new Date();
  return (
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate()
  );
}

function isThisWeek(value) {
  const d = new Date(value);
  const now = new Date();
  // Week starts on Sunday.
  const startOfWeek = new Date(now.getFullYear(), now.getMonth(), now.getDate() - now.getDay());
  return d >= startOfWeek;
}

const filteredIssues = computed(() => {
  switch (activeFilter.value) {
    case "created-today":
      return issues.value.filter((i) => isToday(i.created_at));
    case "completed-today":
      return issues.value.filter((i) => i.completed_at && isToday(i.completed_at));
    case "created-week":
      return issues.value.filter((i) => isThisWeek(i.created_at));
    case "completed-week":
      return issues.value.filter((i) => i.completed_at && isThisWeek(i.completed_at));
    default:
      return issues.value;
  }
});

const issuesByStatus = computed(() => {
  const groups = Object.fromEntries(columns.map((c) => [c.status, []]));
  for (const issue of filteredIssues.value) {
    (groups[issue.status] ?? (groups[issue.status] = [])).push(issue);
  }
  return groups;
});

function toggleFilter(key) {
  activeFilter.value = activeFilter.value === key ? null : key;
}

async function reloadIssues() {
  const res = await fetch("/api/issues");
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  issues.value = await res.json();
}

onMounted(async () => {
  try {
    await reloadIssues();
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
  if (frozen.value) return;

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

function openEdit(issue) {
  if (frozen.value) return;
  editingIssueId.value = issue.id;
}

function closeEdit() {
  editingIssueId.value = null;
}

async function deleteIssue(issue) {
  if (frozen.value) return;
  if (!window.confirm(`Delete issue #${issue.id} "${issue.title}"? This cannot be undone.`)) {
    return;
  }

  frozen.value = true;
  actionError.value = null;
  try {
    const res = await fetch(`/api/issues/${issue.id}`, { method: "DELETE" });
    if (!res.ok) {
      const body = await res.json().catch(() => null);
      throw new Error(body?.message || `${res.status} ${res.statusText}`);
    }
    if (expandedId.value === issue.id) expandedId.value = null;
    await reloadIssues();
  } catch (e) {
    actionError.value = e.message;
  } finally {
    frozen.value = false;
  }
}

async function onIssueUpdated() {
  editingIssueId.value = null;
  // Subtasks may have changed; drop the cache so a re-expand refetches.
  subtasksById.value = {};
  expandedId.value = null;

  try {
    await reloadIssues();
  } catch (e) {
    actionError.value = e.message;
  }
}

function onDragStart(event, issue) {
  if (frozen.value) {
    event.preventDefault();
    return;
  }
  draggedId.value = issue.id;
  event.dataTransfer.effectAllowed = "move";
  event.dataTransfer.setData("text/plain", String(issue.id));
}

function onDragEnter(status) {
  if (frozen.value || draggedId.value == null) return;
  dragOverStatus.value = status;
}

function onDragLeave(status) {
  if (dragOverStatus.value === status) {
    dragOverStatus.value = null;
  }
}

async function onDrop(event, newStatus) {
  dragOverStatus.value = null;

  const id = draggedId.value;
  draggedId.value = null;
  if (frozen.value || id == null) return;

  const issue = issues.value.find((i) => i.id === id);
  if (!issue || issue.status === newStatus) return;

  frozen.value = true;
  actionError.value = null;
  try {
    const res = await fetch(`/api/issues/${id}/status`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status: newStatus }),
    });
    if (!res.ok) {
      const body = await res.json().catch(() => null);
      throw new Error(body?.message || `${res.status} ${res.statusText}`);
    }
    await reloadIssues();
  } catch (e) {
    actionError.value = e.message;
  } finally {
    frozen.value = false;
  }
}
</script>

<template>
  <main class="mx-auto max-w-6xl p-6">
    <h1 class="mb-4 text-2xl font-semibold text-gray-900">Issues</h1>

    <div class="mb-6 flex flex-wrap gap-2">
      <button
        v-for="f in filterOptions"
        :key="f.key"
        type="button"
        class="rounded-full border px-3 py-1 text-xs font-medium transition-colors"
        :class="
          activeFilter === f.key
            ? 'border-blue-600 bg-blue-600 text-white'
            : 'border-gray-300 bg-white text-gray-600 hover:bg-gray-50'
        "
        @click="toggleFilter(f.key)"
      >
        {{ f.label }}
      </button>
    </div>

    <p v-if="error" class="text-red-600">Failed to load issues: {{ error }}</p>
    <p v-if="actionError" class="mb-4 text-sm text-red-600">{{ actionError }}</p>

    <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
      <section
        v-for="col in columns"
        :key="col.status"
        class="flex flex-col rounded-lg border border-gray-200 border-t-4 bg-gray-50 transition-colors"
        :class="[col.accent, dragOverStatus === col.status ? 'bg-blue-50 ring-2 ring-inset ring-blue-300' : '']"
        @dragover.prevent
        @dragenter.prevent="onDragEnter(col.status)"
        @dragleave="onDragLeave(col.status)"
        @drop="onDrop($event, col.status)"
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
              class="rounded-full bg-blue-600 px-2 py-0.5 text-xs font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              :disabled="frozen"
              @click="showNewIssueModal = true"
            >
              + Add
            </button>
          </div>
        </header>

        <div class="flex min-h-16 flex-col gap-3 px-3 pb-3">
          <article
            v-for="issue in issuesByStatus[col.status]"
            :key="issue.id"
            :draggable="!frozen"
            class="cursor-pointer rounded-lg border border-gray-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md"
            :class="{ 'opacity-50': draggedId === issue.id }"
            @dragstart="onDragStart($event, issue)"
            @dragend="draggedId = null"
            @click="toggleExpanded(issue.id)"
          >
            <div class="mb-2 flex items-center justify-between">
              <span class="font-mono text-xs text-gray-400">#{{ issue.id }}</span>
              <div class="flex items-center gap-1">
                <button
                  type="button"
                  class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 disabled:opacity-50"
                  :disabled="frozen"
                  title="Edit issue"
                  @click.stop="openEdit(issue)"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                    <path
                      d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793 3 14.172V17h2.828l8.38-8.379-2.83-2.828z"
                    />
                  </svg>
                </button>
                <button
                  type="button"
                  class="rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-600 disabled:opacity-50"
                  :disabled="frozen"
                  title="Delete issue"
                  @click.stop="deleteIssue(issue)"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                    <path
                      fill-rule="evenodd"
                      d="M8.75 1A2.75 2.75 0 006 3.75v.443c-.795.077-1.584.176-2.365.298a.75.75 0 10.23 1.482l.149-.022.841 10.518A2.75 2.75 0 007.596 19h4.808a2.75 2.75 0 002.742-2.53l.841-10.52.149.023a.75.75 0 00.23-1.482A41.03 41.03 0 0014 4.193V3.75A2.75 2.75 0 0011.25 1h-2.5zM10 4c.84 0 1.673.025 2.5.075V3.75c0-.69-.56-1.25-1.25-1.25h-2.5c-.69 0-1.25.56-1.25 1.25v.325C8.327 4.025 9.16 4 10 4zM8.58 7.72a.75.75 0 00-1.5.06l.3 7.5a.75.75 0 101.5-.06l-.3-7.5zm4.34.06a.75.75 0 10-1.5-.06l-.3 7.5a.75.75 0 101.5.06l.3-7.5z"
                      clip-rule="evenodd"
                    />
                  </svg>
                </button>
              </div>
            </div>
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

    <EditIssueModal
      v-if="editingIssueId !== null"
      :issue-id="editingIssueId"
      @close="closeEdit"
      @updated="onIssueUpdated"
      @freeze="frozen = $event"
    />

    <div v-if="frozen" class="fixed inset-0 z-40 flex items-center justify-center bg-white/50">
      <svg class="h-10 w-10 animate-spin text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path
          class="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        ></path>
      </svg>
    </div>
  </main>
</template>
