<template>
  <div class="bookshelf">
    <label class="upload-card" v-if="!loading">
      <input type="file" accept=".epub" hidden @change="handleUpload" />
      <div class="upload-icon">+</div>
      <div class="upload-label">Add Book</div>
    </label>
    <div v-if="loading" class="loading">Browsing the shelves&hellip;</div>
    <div v-else-if="filtered.length === 0 && books.length === 0" class="empty-state">
      <div class="icon">&#128218;</div>
      <p>
        Your library is empty.<br />
        <small>Drop EPUBs via Finder or click + to upload.</small>
      </p>
    </div>
    <div v-else-if="filtered.length === 0" class="empty-state">
      <p>No books match "{{ query }}".</p>
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

async function handleUpload(e) {
  const file = e.target.files?.[0]
  if (!file) return
  loading.value = true
  try {
    const form = new FormData()
    form.append('file', file)
    const res = await fetch('/api/books/upload', { method: 'POST', body: form })
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      showToast(data.error || 'Upload failed')
    } else {
      showToast(`Added "${file.name}"`)
    }
  } catch {
    showToast('Upload failed')
  }
  await fetchBooks()
  e.target.value = ''
}

async function handleDelete(id) {
  await fetch(`/api/books/${id}`, { method: 'DELETE' })
  books.value = books.value.filter((b) => b.id !== id)
  showToast('Book removed from library')
}

onMounted(fetchBooks)
</script>
