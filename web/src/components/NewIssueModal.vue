<script setup>
import { nextTick, ref } from "vue";

const emit = defineEmits(["close", "created"]);

const title = ref("");
const description = ref("");
const subtasks = ref([""]);
const subtaskInputs = ref([]);
const submitting = ref(false);
const error = ref(null);

function addSubtask() {
  subtasks.value.push("");
}

function removeSubtask(index) {
  subtasks.value.splice(index, 1);
}

function onSubtaskEnter(index) {
  subtasks.value.splice(index + 1, 0, "");
  nextTick(() => {
    subtaskInputs.value[index + 1]?.focus();
  });
}

async function submit() {
  if (!title.value.trim()) {
    error.value = "Title is required";
    return;
  }

  error.value = null;
  submitting.value = true;
  try {
    const res = await fetch("/api/issues", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        title: title.value,
        description: description.value,
        subtasks: subtasks.value.map((s) => s.trim()).filter(Boolean),
      }),
    });

    if (!res.ok) {
      const body = await res.json().catch(() => null);
      throw new Error(body?.message || `${res.status} ${res.statusText}`);
    }

    emit("created", await res.json());
  } catch (e) {
    error.value = e.message;
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
    @click.self="emit('close')"
  >
    <div class="w-full max-w-lg rounded-lg bg-white p-6 shadow-xl">
      <h2 class="mb-4 text-lg font-semibold text-gray-900">New Issue</h2>

      <form class="flex flex-col gap-4" @submit.prevent="submit">
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Title</label>
          <input
            v-model="title"
            type="text"
            class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-400 focus:outline-none"
            placeholder="Issue title"
          />
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Description</label>
          <textarea
            v-model="description"
            rows="3"
            class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-400 focus:outline-none"
            placeholder="Describe the issue"
          ></textarea>
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">Subtasks</label>
          <div class="flex flex-col gap-2">
            <div v-for="(_, index) in subtasks" :key="index" class="flex gap-2">
              <input
                :ref="(el) => (subtaskInputs[index] = el)"
                v-model="subtasks[index]"
                type="text"
                class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-400 focus:outline-none"
                placeholder="Subtask"
                @keydown.enter.prevent="onSubtaskEnter(index)"
              />
              <button
                type="button"
                class="rounded-md border border-gray-300 px-2 text-gray-500 hover:bg-gray-50"
                @click="removeSubtask(index)"
              >
                &times;
              </button>
            </div>
          </div>
          <button
            type="button"
            class="mt-2 text-sm font-medium text-blue-600 hover:text-blue-700"
            @click="addSubtask"
          >
            + Add subtask
          </button>
        </div>

        <p v-if="error" class="text-sm text-red-600">{{ error }}</p>

        <div class="mt-2 flex justify-end gap-2">
          <button
            type="button"
            class="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
            @click="emit('close')"
          >
            Cancel
          </button>
          <button
            type="submit"
            :disabled="submitting"
            class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {{ submitting ? "Creating…" : "Create" }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
