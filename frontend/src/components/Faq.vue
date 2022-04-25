<script>
import Modal from "./Modal.vue";
export default {
  components: {
    Modal,
  },
  props: {
    info: Object,
  },
};
</script>

<template>
  <modal title="FAQ">
    <div class="bg-gray-200 p-5 my-3">
      <h5 class="text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">How to get a refund?</h5>
      <p>
        All entries on January 2022 is available for refund. Please send a message in contact form
        with your txid, total amount and monero address to send the refund.
      </p>
    </div>
    <div class="bg-gray-200 p-5">
      <h5 class="text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">When is the draw date?</h5>
      <p>Every first day of the month.</p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <h5 class="text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">How is the entry pot distributed?</h5>
      <p>
        70% to the winner, 20% goes to referral or the next draw if no referral
        and last 10% to server maintenance.
      </p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <h5 class="text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">What is the entry price?</h5>
      <p>
        Entry price is currently {{ info.entry_price }} XMR. It's the value of
        XMR at $2.50 (USD) at the time of adjustment and will be that price in
        XMR for the next 24 hours.
      </p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <h5 class="text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">How can I see my entries?</h5>
      <p>
        Enter your monero address in the address field. Click View Entry Address
        then click the Entries # link.
      </p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <h5 class="text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">How do I know the drawing is fair?</h5>
      <p>
        The drawing of this system is easily verifyable with 2 things. First is
        our sign key, which is the first block of the month. The sign key used
        to create your entry hash using SHA256(SignKey+ID). The resulting hash
        is shown beside your entry id. Entry id is an incrementing number from 1
        to total number of entries. This will always be a deterministic hash
        based on the current drawing month.
      </p>
      <p class="mt-2">
        After the month, we will get the first block hash of the next month and
        use this to compare all previous entries to select the winning entries.
        Entries with similar total characters are equally splitting the winning
        pot.
      </p>
      <p class="mt-2">
        Only characters that aligns with your hash to block hash are counted. So
        "abc123" and "aaa1bb" means 2 matches.
      </p>
      <p class="mt-2">
        You can use this
        <a href="https://emn178.github.io/online-tools/sha256.html" class="underline" target="_blank">SHA256 tool</a>
        to verify that your hashes are valid.
      </p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <h5 class="text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">Why do you ask for a username?</h5>
      <p>
        Username will keep your monero entry address for the next drawing if you
        have entries or referred an entry.
      </p>
      <p class="mt-2">
        With username you can use it for affiliate and earn 20% for each entry
        you refer. It’s paid on the next drawing. If your referral amount is
        less than the entry price at the time of drawing you'll be credited in
        your monero address where you can add additional XMR to make it a valid
        entry.
      </p>
    </div>
    <div class="bg-gray-200 p-5 my-3">
      <h5 class="text-lg font-semibold border-b-2 border-gray-400 pb-2 mb-4">I want to increase the pot without entries,
        what do I do?</h5>
      <p>You can transfer XMR to the address below:</p>
      <img :src="'/api/internal/QrCode?addr=' + info.address" />
      <input type="text" class="w-full py-2 my-1 bg-gray-200" readonly :value="info.address" />
    </div>
  </modal>
</template>