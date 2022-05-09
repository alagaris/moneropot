<script>
import axios from "axios";
import TextHightlight from "./TextHighlight.vue"
import Modal from "./Modal.vue";
import sha256 from "js-sha256"

const fistMonth = "2022-02";
const cache = {};

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
      const amount = this.winInfo.amount + "";
      const l = amount.length;
      return amount.substr(0, l - 12) + "." + amount.substr(l - 12);
    },
    multiWinner() {
      return Object.keys(this.winInfo.accounts).length > 0;
    },
    winningEntries() {
      let entries = [];
      for (let key in this.winInfo.accounts) {
        entries = entries.concat(this.winInfo.accounts[key]);
      }
      return entries;
    },
    winInfo() {
      if (this.winner) {
        return this.winner;
      }
      return this.info.last_winner;
    },
    hasPrev() {
      return this.winInfo.date > fistMonth
    },
    hasNext() {
      return this.winInfo.date < this.info.last_winner.date
    }
  },
  methods: {
    sha256(text) {
      return sha256(text)
    },
    calcMatches() {
      this.matchInfo = {
        hash: sha256(this.winInfo.sign_key + this.testId),
        indexes: [],
        count: 0
      };
      for (let i = 0; i < this.winInfo.block.length; i++) {
        if (this.winInfo.block[i] === this.matchInfo.hash[i]) {
          this.matchInfo.count++;
          this.matchInfo.indexes.push(i);
        }
      }
    },
    setWinner(np = 0) {
      const dt = new Date(this.winInfo.date + "-15");
      let year = dt.getFullYear();
      let month = dt.getMonth() + np;
      if (month <= 0) {
        year--;
        month = 12;
      }
      if (month < 10) {
        month = "0" + month;
      }
      const key = year + "-" + month
      if (cache[key]) {
        this.winner = cache[key];
        this.calcMatches();
      } else {
        axios.get("/api/internal/Winner?dt=" + key).then(resp => {
          this.winner = resp.data;
          cache[key] = resp.data
          this.calcMatches();
        });
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
      matchInfo: null,
      winner: null
    }
  }
};
</script>

<template>
  <modal title="Winner">
    <div>
      <span class="font-semibold p-2">Date</span>
      <a class="underline cursor-pointer" @click="setWinner(0)" v-if="hasPrev">&lt;&lt;</a>
      {{ winInfo.date }}
      <a class="underline cursor-pointer" @click="setWinner(2)" v-if="hasNext">&gt;&gt;</a>
    </div>
    <div>
      <span class="font-semibold p-2">Sign Key</span>
      <a target="_blank" :href="'https://xmrchain.net/block/' + winInfo.sign_key" class="underline">{{
          winInfo.sign_key
      }}</a>
    </div>
    <div class>
      <span class="font-semibold p-2">First Block</span>
      <a target="_blank" :href="'https://xmrchain.net/block/' + winInfo.block" class="underline">
        {{ winInfo.block }}
      </a>
    </div>
    <div class>
      <span class="font-semibold p-2">Entries</span>
      {{ winInfo.entries }}
    </div>
    <div class>
      <span class="font-semibold p-2">Winnings</span>
      {{ xmrWinnings }}
      <small class="text-sm text-gray-500">({{ totalWinnings }})</small>
    </div>
    <div class="my-4">
      <span class="font-bold p-2">Winner{{ multiWinner ? "s" : "" }}</span>
      <div class v-for="(entries, addr) in winInfo.accounts" :key="addr">
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
            <td class="font-mono">{{ sha256(winInfo.sign_key + entry) }}</td>
          </tr>
        </tbody>
      </table>
      <label class="pt-3">
        Check Your ID
        <input type="number" class="px-4 py-2 w-24 bg-gray-100" v-model.number="testId" />
      </label>
      <span class="rounded bg-yellow-200 p-2 m-1" v-if="matchInfo">Matches: {{ matchInfo.count }}</span>
      <div v-if="matchInfo">
        <div>
          <text-hightlight :value="winInfo.block" :indexes="matchInfo.indexes" />
        </div>
        <text-hightlight :value="matchInfo.hash" :indexes="matchInfo.indexes" />
      </div>
    </div>
  </modal>
</template>