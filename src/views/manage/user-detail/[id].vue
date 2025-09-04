<script setup lang="ts">
interface Props {
  id: string;
}

defineProps<Props>();
</script>

<template>
  <LookForward />
  {{ id }}
</template>

<style scoped></style>
