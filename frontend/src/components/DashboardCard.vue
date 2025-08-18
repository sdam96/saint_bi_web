<template>
  <div class="col">
    <div class="card h-100">
      <div class="card-header fw-bold">{{ title }}</div>
      <ul class="list-group list-group-flush">
        <slot></slot>
      </ul>
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: String, // La tarjeta acepta una propiedad 'title'
});
</script>