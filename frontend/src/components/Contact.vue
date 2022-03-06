<script>
import axios from "axios";
import Modal from "./Modal.vue";

export default {
  components: { Modal },
  data() {
    return {
      form: {
        contact: null,
        message: null,
      },
      sent: false,
    };
  },
  methods: {
    submitForm() {
      axios.post("/api/internal/Contact", this.form).then(() => {
        this.sent = true;
      });
    },
  },
};
</script>

<template>
  <modal title="Contact">
    <p v-if="sent" class="text-center p-4 bg-green-200 rounded">
      Message sent!
    </p>
    <form v-else class="mt-4" @submit.prevent="submitForm">
      <div class="my-5 text-sm">
        <label for="contact" class="block text-black">Contact Info</label>
        <input
          id="contact"
          type="text"
          autoFocus
          class="
            rounded-sm
            px-4
            py-3
            mt-1
            focus:outline-none
            bg-gray-100
            w-full
          "
          v-model="form.contact"
          placeholder="Enter a form of contact"
        />
        <small class="text-gray-500"
          >Enter any contact info you expect a response</small
        >
      </div>
      <div class="my-5 text-sm">
        <label for="message" class="block text-black"> Message </label>
        <textarea
          id="message"
          v-model="form.message"
          type="text"
          autoFocus
          class="
            rounded-sm
            px-4
            py-3
            mt-1
            focus:outline-none
            bg-gray-100
            w-full
          "
          rows="4"
          placeholder="Questions, suggestions or concerns."
        />
      </div>
      <div class="my-5">
        <button
          :disabled="!form.message"
          type="submit"
          :class="{ 'opacity-50 cursor-not-allowed': !form.message }"
          class="
            rounded-sm
            block
            text-center text-white
            bg-gray-800
            p-3
            duration-300
            hover:bg-black
            w-full
          "
        >
          Send
        </button>
      </div>
    </form>
  </modal>
</template>