<script>
import TextHightlight from "./TextHighlight.vue"
import Modal from "./Modal.vue";
import sha256 from "js-sha256"

export default {
  components: { Modal, TextHightlight },
  props: {
    info: Object,
  },
  computed: {
    totalWinnings() {
      return this.appUtil.toUsd(
        this.info.xmr_rate,
        this.xmrWinnings
      );
    },
    xmrWinnings() {
      const amount = this.info.last_winner.amount + "";
      const l = amount.length;
      return amount.substr(0, l - 12) + "." + amount.substr(l - 12);
    },
    multiWinner() {
      return Object.keys(this.info.last_winner.accounts).length > 0;
    },
    winningEntries() {
      let entries = [];
      for (let key in this.info.last_winner.accounts) {
        entries = entries.concat(this.info.last_winner.accounts[key]);
      }
      return entries;
    }
  },
  methods: {
    sha256(text) {
      return sha256(text)
    },
    calcMatches() {
      this.matchInfo = {
        hash: sha256(this.info.last_winner.sign_key + this.testId),
        indexes: [],
        count: 0
      };
      for (let i = 0; i < this.info.last_winner.block.length; i++) {
        if (this.info.last_winner.block[i] === this.matchInfo.hash[i]) {
          this.matchInfo.count++;
          this.matchInfo.indexes.push(i);
        }
      }
    }
  },
  watch: {
    testId() {
      this.matchInfo = null;
      if (this.testId && typeof this.testId === "number") {
        this.calcMatches();
      }
    }
  },
  data() {
    return {
      testId: null,
      matchInfo: null
    }
  }
};
</script>

<template>
  <modal title="Winner">
    <div class>
      <span class="font-semibold p-2">Sign Key</span>
      <a
        target="_blank"
        :href="'https://xmrchain.net/block/' + info.last_winner.sign_key"
        class="underline"
      >{{ info.last_winner.sign_key }}</a>
    </div>
    <div class>
      <span class="font-semibold p-2">First Block</span>
      <a
        target="_blank"
        :href="'https://xmrchain.net/block/' + info.last_winner.block"
        class="underline"
      >
        <template v-if="matchInfo">
          <text-hightlight :value="info.last_winner.block" :indexes="matchInfo.indexes" />
        </template>
        <template v-else>{{ info.last_winner.block }}</template>
      </a>
    </div>
    <div class>
      <span class="font-semibold p-2">Entries</span>
      {{ info.last_winner.entries }}
    </div>
    <div class>
      <span class="font-semibold p-2">Winnings</span>
      {{ xmrWinnings }}
      <small class="text-sm text-gray-500">({{ totalWinnings }})</small>
    </div>
    <div class="my-4">
      <span class="font-bold p-2">Winner{{ multiWinner ? "s" : "" }}</span>
      <div class v-for="(entries, addr) in info.last_winner.accounts" :key="addr">
        <div class>
          <span class="font-semibold p-2">{{ addr }}</span>
          {{ entries }}
        </div>
      </div>
      <table class="table-auto w-full mb-4">
        <thead>
          <tr>
            <th class="p-2">ID</th>
            <th class="p-2">HASH</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="entry in winningEntries" :key="entry">
            <td class="underline cursor-pointer" @click="testId = entry">{{ entry }}</td>
            <td>{{ sha256(info.last_winner.sign_key + entry) }}</td>
          </tr>
        </tbody>
      </table>
      <label class="pt-3">
        Check Your ID
        <input
          type="number"
          class="px-4 py-2 w-24 bg-gray-100"
          v-model.number="testId"
        />
      </label>
      <span class="rounded bg-yellow-200 p-2 m-1" v-if="matchInfo">Matches: {{ matchInfo.count }}</span>
      <div v-if="matchInfo">
        <text-hightlight :value="matchInfo.hash" :indexes="matchInfo.indexes" />
      </div>
    </div>
  </modal>
</template>