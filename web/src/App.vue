<template>
  <AppHeader :query="query" @update:query="query = $event" />
  <main class="main">
    <Dashboard :query="query" />
  </main>
</template>

<script setup>
import { ref } from 'vue'
import AppHeader from './components/AppHeader.vue'
import Dashboard from './views/Dashboard.vue'

const query = ref('')
</script>
