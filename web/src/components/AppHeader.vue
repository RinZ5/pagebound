<template>
  <header class="header">
    <h1>Pagebound</h1>
    <em>a cozy collection</em>
    <input
      class="search-bar"
      type="text"
      :value="query"
      @input="$emit('update:query', $event.target.value)"
      placeholder="Search by title or author..."
    />
  </header>
</template>

<script setup>
defineProps({ query: { type: String, default: '' } })
defineEmits(['update:query'])
</script>
