<script>
export default {
  props: {
    title: String,
    modalId: String,
  },
  methods: {
    toggleModal(payload) {
      this.$refs.dialog.classList.toggle("opacity-0");
      this.$refs.dialog.classList.toggle("pointer-events-none");
      if (this.visible) {
        setTimeout(() => {
          this.visible = false;
        }, 250);
      } else {
        this.eventBus.emit("modal-show-" + this.modId, payload);
        this.visible = true;
      }
    },
  },
  computed: {
    modId() {
      if (!this.modalId) {
        return this.title ? this.title.toLowerCase() : this._uid;
      }
      return this.modalId;
    },
  },
  data() {
    return {
      visible: false,
    };
  },
  created() {
    this.eventBus.on("toggle-modal-" + this.modId, this.toggleModal);
  },
};
</script>

<style scoped>
.modal {
  transition: opacity 0.25s ease;
}
</style>

<template>
  <!--Modal-->
  <div
    ref="dialog"
    class="
      modal
      opacity-0
      pointer-events-none
      fixed
      w-full
      h-full
      top-0
      left-0
      flex
      items-center
      justify-center
      z-50
      py-10
    "
  >
    <div
      class="modal-overlay absolute w-full h-full bg-gray-900 opacity-50"
    ></div>

    <div
      class="
        modal-container
        bg-white
        w-11/12
        sm:w-6/12
        mx-auto
        shadow-lg
        z-50
        max-h-full
        overflow-y-auto
      "
    >
      <div
        class="
          modal-close
          absolute
          top-0
          right-0
          cursor-pointer
          flex flex-col
          items-center
          mt-4
          mr-4
          text-white text-sm
          z-50
        "
      >
        <svg
          class="fill-current text-white"
          xmlns="http://www.w3.org/2000/svg"
          width="18"
          height="18"
          viewBox="0 0 18 18"
          @click="toggleModal"
        >
          <path
            d="M14.53 4.53l-1.06-1.06L9 7.94 4.53 3.47 3.47 4.53 7.94 9l-4.47 4.47 1.06 1.06L9 10.06l4.47 4.47 1.06-1.06L10.06 9z"
          ></path>
        </svg>
      </div>
      <!-- Add margin if you want to see some of the overlay behind the modal-->
      <div class="modal-content py-4 text-left px-6">
        <!--Title-->
        <div class="flex justify-between items-center pb-3">
          <p class="text-2xl font-bold">{{ title }}</p>
          <div class="modal-close cursor-pointer z-50">
            <svg
              class="fill-current text-black"
              xmlns="http://www.w3.org/2000/svg"
              width="18"
              height="18"
              viewBox="0 0 18 18"
              @click="toggleModal"
            >
              <path
                d="M14.53 4.53l-1.06-1.06L9 7.94 4.53 3.47 3.47 4.53 7.94 9l-4.47 4.47 1.06 1.06L9 10.06l4.47 4.47 1.06-1.06L10.06 9z"
              ></path>
            </svg>
          </div>
        </div>
        <slot v-if="visible" />
      </div>
    </div>
  </div>
</template>