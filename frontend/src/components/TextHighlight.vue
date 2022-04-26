<script>
export default {
  props: {
    value: String,
    indexes: Array,
    highlight: {
      type: String,
      default: "bg-yellow-200"
    }
  }
}
</script>

<template>
  <span class="font-mono" v-for="(v, idx) in value" :key="idx"
    :class="[indexes.indexOf(idx) !== -1 ? highlight : null]">{{ v }}</span>
</template>