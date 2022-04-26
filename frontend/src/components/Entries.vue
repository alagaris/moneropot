<script>
import axios from "axios";
import Spinner from "./Spinner.vue";
import Modal from "./Modal.vue";
import sha256 from "js-sha256"

export default {
  components: {
    Modal,
    Spinner,
  },
  props: {
    info: Object,
  },
  data() {
    return {
      entries: null,
      page: 1,
      limit: 100,
      accountId: null,
      hasMore: false,
      loading: false,
    };
  },
  methods: {
    sha256(text) {
      return sha256(text);
    },
    loadEntries(reset) {
      this.loading = true;
      if (reset) {
        this.entries = null;
        this.page = 1;
      } else {
        this.page++;
      }
      axios
        .get("/api/entries?p=" + this.page + "&a=" + (this.accountId || ""))
        .then((resp) => {
          const entries = resp.data;
          if (this.entries === null) {
            this.entries = [];
          }
          if (entries !== null) {
            this.entries = this.entries.concat(entries);
            this.hasMore = entries.length >= this.limit;
          } else {
            this.hasMore = false;
          }
          this.loading = false;
        });
    },
  },
  created() {
    this.eventBus.on("modal-show-entries", (accountId) => {
      this.accountId = accountId;
      if (this.accountId) {
        this.loadEntries(true);
      } else {
        this.entries = [];
        this.hasMore = false;
        for (let i = 1; i <= this.info.entries; i++) {
          this.entries.push({ id: i, hash: sha256(this.info.sign_key + i) })
        }
      }
    });
  },
};
</script>

<template>
  <modal modal-id="entries" :title="accountId ? 'Your Entries' : 'All Entries (' + info.entries + ')'">
    <span class="text-xs">
      SignKey
      <a target="_blank" href="https://emn178.github.io/online-tools/sha256.html">SHA256</a>(
      <a target="_blank" :href="'https://xmrchain.net/block/' + info.sign_key" class="underline">{{ info.sign_key
      }}</a>&nbsp;+&nbsp;
      <span class="font-bold">ID</span>)
    </span>
    <table class="table-auto w-full">
      <thead>
        <tr>
          <th class="p-2">ID</th>
          <th class="p-2">HASH</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(entry, index) in entries" :key="index">
          <td class="py-1 px-2">{{ entry.id }}</td>
          <td class="py-1 px-2 font-mono">{{ entry.hash }}</td>
        </tr>
      </tbody>
    </table>
    <div class="text-center" v-if="hasMore">
      <spinner v-if="loading" />
      <button v-else type="button" class="bg-gray-500 px-3 py-2 text-white rounded-sm mt-4"
        @click="loadEntries(false)">View More</button>
    </div>
  </modal>
</template>