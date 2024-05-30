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
      <a href="https://reddit.com/r/monero" target="_blank"
        class="block text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">r/Monero</a>
      <p>Reddit forum for monero.</p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <a href="https://monero.observer/" target="_blank"
        class="block text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">Monero Observer</a>
      <p>Everything that moves in the Monero community.</p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <a href="https://www.revuo-xmr.com/" target="_blank"
        class="block text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">Revuo Monero</a>
      <p>Monero updates.</p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <a href="https://ccs.getmonero.org/" target="_blank"
        class="block text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">Community Crowdfunding System</a>
      <p>Support monero development by donating to General Fund or CCS proposals.</p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <a href="https://monero.com/" target="_blank"
        class="block text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">Monero.com</a>
      <p>Buy, use, sell, and learn about Monero, the leading private digital payment method.</p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <a href="https://featherwallet.org/" target="_blank"
        class="block text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">Feather: A Free Monero Desktop
        Wallet</a>
      <p>Feather is a free, open-source Monero wallet for Linux, Tails, Windows and macOS.</p>
    </div>
    <p class="text-center p-1 text-xs">Have a suggestion for a new link? Send a message in contact form.</p>
  </modal>
</template>