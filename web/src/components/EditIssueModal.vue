<script setup>
import { onMounted, ref } from "vue";

const props = defineProps({
  issueId: { type: Number, required: true },
});

const emit = defineEmits(["close", "updated", "freeze"]);

const loading = ref(true);
const loadError = ref(null);
const submitting = ref(false);
const submitError = ref(null);

const title = ref("");
const description = ref("");
const status = ref("open");
const assignedTo = ref("");
const subtasks = ref([]);
const resolutionNotes = ref("");

onMounted(async () => {
  try {
    const res = await fetch(`/api/issues/${props.issueId}`);
    if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
    const detail = await res.json();

    title.value = detail.title;
    description.value = detail.description;
    status.value = detail.status;
    assignedTo.value = detail.assigned_to;
    subtasks.value = (detail.subtasks ?? []).map((s) => ({ text: s.text, done: s.done }));
    resolutionNotes.value = detail.resolution_notes;
  } catch (e) {
    loadError.value = e.message;
  } finally {
    loading.value = false;
  }
});

function addSubtask() {
  subtasks.value.push({ text: "", done: false });
}

function removeSubtask(index) {
  subtasks.value.splice(index, 1);
}

async function submit() {
  if (!title.value.trim()) {
    submitError.value = "Title is required";
    return;
  }

  submitError.value = null;
  submitting.value = true;
  emit("freeze", true);
  try {
    const res = await fetch(`/api/issues/${props.issueId}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        title: title.value,
        description: description.value,
        status: status.value,
        assigned_to: assignedTo.value,
        subtasks: subtasks.value
          .map((s) => ({ text: s.text.trim(), done: s.done }))
          .filter((s) => s.text),
        resolution_notes: resolutionNotes.value,
      }),
    });

    if (!res.ok) {
      const body = await res.json().catch(() => null);
      throw new Error(body?.message || `${res.status} ${res.statusText}`);
    }

    emit("updated");
  } catch (e) {
    submitError.value = e.message;
  } finally {
    submitting.value = false;
    emit("freeze", false);
  }
}
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
    @click.self="!submitting && emit('close')"
  >
    <div class="w-full max-w-lg rounded-lg bg-white p-6 shadow-xl">
      <h2 class="mb-4 text-lg font-semibold text-gray-900">Edit Issue #{{ issueId }}</h2>

      <p v-if="loading" class="text-sm text-gray-400 italic">Loading…</p>
      <p v-else-if="loadError" class="text-sm text-red-600">Failed to load issue: {{ loadError }}</p>

      <form v-else class="flex flex-col gap-4" @submit.prevent="submit">
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Title</label>
          <input
            v-model="title"
            type="text"
            :disabled="submitting"
            class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-400 focus:outline-none disabled:opacity-50"
            placeholder="Issue title"
          />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">Status</label>
            <select
              v-model="status"
              :disabled="submitting"
              class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-400 focus:outline-none disabled:opacity-50"
            >
              <option value="open">Open</option>
              <option value="in-progress">In Progress</option>
              <option value="done">Done</option>
            </select>
          </div>

          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">Assigned To</label>
            <input
              v-model="assignedTo"
              type="text"
              :disabled="submitting"
              class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-400 focus:outline-none disabled:opacity-50"
              placeholder="Unassigned"
            />
          </div>
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Description</label>
          <textarea
            v-model="description"
            rows="3"
            :disabled="submitting"
            class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-400 focus:outline-none disabled:opacity-50"
            placeholder="Describe the issue"
          ></textarea>
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Subtasks</label>
          <div class="flex flex-col gap-2">
            <div v-for="(subtask, index) in subtasks" :key="index" class="flex items-center gap-2">
              <input
                v-model="subtask.done"
                type="checkbox"
                :disabled="submitting"
                class="h-4 w-4 shrink-0 rounded border-gray-300 disabled:opacity-50"
              />
              <input
                v-model="subtask.text"
                type="text"
                :disabled="submitting"
                class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-400 focus:outline-none disabled:opacity-50"
                placeholder="Subtask"
              />
              <button
                type="button"
                :disabled="submitting"
                class="rounded-md border border-gray-300 px-2 text-gray-500 hover:bg-gray-50 disabled:opacity-50"
                @click="removeSubtask(index)"
              >
                &times;
              </button>
            </div>
          </div>
          <button
            type="button"
            :disabled="submitting"
            class="mt-2 text-sm font-medium text-blue-600 hover:text-blue-700 disabled:opacity-50"
            @click="addSubtask"
          >
            + Add subtask
          </button>
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Resolution Notes</label>
          <textarea
            v-model="resolutionNotes"
            rows="2"
            :disabled="submitting"
            class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-400 focus:outline-none disabled:opacity-50"
            placeholder="Notes about how this was resolved"
          ></textarea>
        </div>

        <p v-if="submitError" class="text-sm text-red-600">{{ submitError }}</p>

        <div class="mt-2 flex justify-end gap-2">
          <button
            type="button"
            :disabled="submitting"
            class="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
            @click="emit('close')"
          >
            Cancel
          </button>
          <button
            type="submit"
            :disabled="submitting"
            class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {{ submitting ? "Saving…" : "Save" }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
