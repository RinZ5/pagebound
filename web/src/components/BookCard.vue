<template>
  <div class="book-card">
    <div class="cover-wrap">
      <img v-if="book.cover_url" :src="book.cover_url" :alt="book.title" loading="lazy" />
      <div v-else class="cover-placeholder">&#128214;</div>
    </div>
    <div class="book-info">
      <div class="book-title">{{ book.title }}</div>
      <div class="book-creator" v-if="book.creator">{{ book.creator }}</div>
      <div class="book-actions">
        <a :href="`/api/books/${book.id}/download`" class="download-btn" download>Download</a>
        <button class="delete-btn" @click="confirmDelete">Delete</button>
      </div>
    </div>
    <div class="shelf"></div>
  </div>
</template>

<script setup>
const props = defineProps({
  book: { type: Object, required: true },
})

const emit = defineEmits(['delete'])

function confirmDelete() {
  if (confirm(`Remove "${props.book.title}" from the library?`)) {
    emit('delete', props.book.id)
  }
}
</script>
