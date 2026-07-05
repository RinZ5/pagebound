<template>
  <div class="bookshelf">
    <div v-if="loading" class="loading">Browsing the shelves&hellip;</div>
    <div v-else-if="filtered.length === 0" class="empty-state">
      <div class="icon">&#128218;</div>
      <p v-if="books.length === 0">
        Your library is empty.<br />
        <small>Connect via Finder and drop in some EPUBs.</small>
      </p>
      <p v-else>No books match "{{ query }}".</p>
    </div>
    <BookCard
      v-for="book in filtered"
      :key="book.id"
      :book="book"
      @delete="handleDelete"
    />
  </div>
  <div v-if="toast" class="toast">{{ toast }}</div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import BookCard from '../components/BookCard.vue'

const props = defineProps({
  query: { type: String, default: '' },
})

const books = ref([])
const loading = ref(true)
const toast = ref('')

const filtered = computed(() => {
  const q = props.query.toLowerCase().trim()
  if (!q) return books.value
  return books.value.filter(
    (b) =>
      b.title.toLowerCase().includes(q) ||
      (b.creator && b.creator.toLowerCase().includes(q))
  )
})

async function fetchBooks() {
  loading.value = true
  try {
    const res = await fetch('/api/books')
    const data = await res.json()
    books.value = data.books || []
  } catch {
    books.value = []
  } finally {
    loading.value = false
  }
}

function showToast(msg) {
  toast.value = msg
  setTimeout(() => (toast.value = ''), 2500)
}

async function handleDelete(id) {
  await fetch(`/api/books/${id}`, { method: 'DELETE' })
  books.value = books.value.filter((b) => b.id !== id)
  showToast('Book removed from library')
}

onMounted(fetchBooks)
</script>
