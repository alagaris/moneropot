<script>
export default {
  props: {
    label: String,
    success: Boolean,
  },
};
</script>

<template>
  <div
    class="relative bg-white p-4 border-r-8 shadow-md my-4 flex justify-between"
    :class="[success ? 'border-gray-600' : 'border-yellow-700']"
  >
    <div>
      <p>{{ label }}</p>
    </div>
    <div>
      <p><slot /></p>
    </div>
  </div>
</template>