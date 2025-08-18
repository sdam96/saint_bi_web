// src/components/DateRangePicker.vue
<template>
  <div class="row justify-content-center mt-4 g-3 align-items-center bg-light p-3 rounded border">
    <div class="col-auto">
      <label for="startDate" class="form-label mb-0 fw-bold">Desde:</label>
    </div>
    <div class="col-md-3">
      <input type="date" id="startDate" class="form-control" :value="startDate" @change="updateStartDate($event.target.value)">
    </div>
    <div class="col-auto">
      <label for="endDate" class="form-label mb-0 fw-bold">Hasta:</label>
    </div>
    <div class="col-md-3">
      <input type="date" id="endDate" class="form-control" :value="endDate" @change="updateEndDate($event.target.value)">
    </div>
  </div>
</template>

<script setup>
defineProps({
  startDate: String,
  endDate: String,
});

const emit = defineEmits(['update:startDate', 'update:endDate']);

const updateStartDate = (value) => {
  emit('update:startDate', value);
};

const updateEndDate = (value) => {
  emit('update:endDate', value);
};
</script>