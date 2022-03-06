<script>
import Modal from "./Modal.vue";
export default {
  components: {
    Modal,
  },
};
</script>

<template>
  <modal modal-id="MoneroRelated" title="Spend / Earn / Get Monero">
    <div class="bg-gray-200 p-5 my-3">
      <a
        href="https://xmrmemes.com/"
        target="_blank"
        class="block text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4"
      >XMR Memes</a>
      <p>Get tipped Monero (XMR) for making memes about Monero.</p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <a
        href="https://monero.win/"
        target="_blank"
        class="block text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4"
      >Monero.win</a>
      <p>Win Monero by guessing what the last digit of the next Monero block hash will be!</p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <a
        href="https://localmonero.co/"
        target="_blank"
        class="block text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4"
      >Local Monero</a>
      <p>Buy and sell Monero (XMR) online without ID verification.</p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <a
        href="https://monero.style/"
        target="_blank"
        class="block text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4"
      >Monero Style</a>
      <p>Show Your Support for Monero with Some Monero Merch</p>
    </div>
    <p
      class="text-center p-1 text-xs"
    >Have a suggestion for a new link? Send a message in contact form.</p>
  </modal>
</template>