<script>
import axios from "axios";
import LineInfo from "./components/LineInfo.vue";
import QRCodeVue3 from "qrcode-vue3";
import Faq from "./components/Faq.vue";
import MoneroRelated from "./components/MoneroRelated.vue"
import Contact from "./components/Contact.vue";
import Entries from "./components/Entries.vue";
import Winner from "./components/Winner.vue";

const apiUrl = "/api/";

const errorMap = {
  invalid: "please provide a valid address",
  username: "must be alphanumeric 3-25 chars",
  exists: "choose a different username",
};

export default {
  name: "App",
  components: {
    LineInfo,
    QRCodeVue3,
    Faq,
    MoneroRelated,
    Contact,
    Entries,
    Winner,
  },
  computed: {
    addressUri() {
      if (this.account) {
        if (this.entryQty !== 1 || Number(this.account.xmr) > 0) {
          let price = Number(this.info.entry_price);
          price = ((price * this.entryQty) - Number(this.account.xmr)).toFixed(12);
          return this.account.address_uri.replace(/(tx_amount)=([\d\\.]+)/, '$1=' + price);
        }
        return this.account.address_uri;
      }
      return null;
    },
    formValid() {
      return (
        this.form.address &&
        this.form.address.length === 95 &&
        !this.info.wallet_offline
      );
    },
    refUrl() {
      return this.account && this.account.username
        ? location.origin + "/" + this.account.username
        : null;
    },
    timers() {
      return {
        until_draw: this.getTimeLeft("until_draw"),
        until_price: this.getTimeLeft("until_price")
      }
    }
  },
  methods: {
    getTimeLeft(key) {
      const elapsedTime = this.currentTime - this.initTime;
      const t = (this.info[key] * 1000) - elapsedTime;
      let timeLeft = "";
      if (!this.currentTime) return timeLeft;
      if (t >= 0) {
        let days = Math.floor(t / (1000 * 60 * 60 * 24));
        let hours = Math.floor((t % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
        let mins = Math.floor((t % (1000 * 60 * 60)) / (1000 * 60));
        let secs = Math.floor((t % (1000 * 60)) / 1000);
        timeLeft =
          (days ? days + " day" + (days != 1 ? "s " : " ") : "") +
          (days || hours ? hours + " hr" + (hours != 1 ? "s " : " ") : "") +
          (days || hours || mins
            ? mins + " min" + (mins != 1 ? "s " : " ")
            : "") +
          (days || hours || mins || secs
            ? secs + " sec" + (secs != 1 ? "s" : "")
            : "");
      } else {
        if (key === "until_draw") {
          timeLeft = "Waiting for first block...";
        } else {
          // location.reload();
        }
      }
      return timeLeft;
    },
    viewEntries(accountId) {
      this.toggleModal("entries", accountId);
    },
    submitForm(event, qty = 1) {
      this.entryQty = qty;
      const form = {};
      if (location.pathname !== "/") {
        form.ref = location.pathname.substr(1);
      }
      for (let key in this.form) {
        if (this.form[key] && this.form[key].trim().length) {
          form[key] = this.form[key].trim();
        }
      }
      this.loading++;
      axios
        .post(apiUrl + "accounts", form)
        .then((r) => {
          this.loading--;
          this.account = r.data;
          this.listenToEvents();
          this.$nextTick(() => {
            if (this.$refs.accountInfo) {
              this.$refs.accountInfo.scrollIntoView({
                behavior: "smooth",
                block: "start",
                inline: "nearest",
              });
            }
          });
        })
        .catch((err) => {
          this.loading--;
          const error = err.response.data;
          if (error.error === "validation") {
            this.error = error.params;
          } else {
            this.error = { message: error.error };
          }
        });
    },
    errorField(field) {
      if (this.error && this.error[field]) {
        return errorMap[this.error[field]] || this.error[field];
      }
      return null;
    },
    toUsd(value) {
      return this.appUtil.toUsd(this.info.xmr_rate, value);
    },
    copyToClipboard(text) {
      if (window.clipboardData && window.clipboardData.setData) {
        // Internet Explorer-specific code path to prevent textarea being shown while dialog is visible.
        return window.clipboardData.setData("Text", text);
      } else if (
        document.queryCommandSupported &&
        document.queryCommandSupported("copy")
      ) {
        var textarea = document.createElement("textarea");
        textarea.textContent = text;
        textarea.style.position = "fixed"; // Prevent scrolling to bottom of page in Microsoft Edge.
        document.body.appendChild(textarea);
        textarea.select();
        try {
          return document.execCommand("copy"); // Security exception may be thrown by some browsers.
        } catch (ex) {
          console.warn("Copy to clipboard failed.", ex);
          return false;
        } finally {
          document.body.removeChild(textarea);
        }
      }
    },
    toggleModal(title, payload) {
      this.eventBus.emit("toggle-modal-" + title, payload);
    },
    listenToEvents() {
      if (!window.EventSource) return;
      if (this.eventSource) {
        try {
          this.eventSource.close();
        } catch (e) {
          console.log("eventSource error", e)
        }
      } else {
        this.loadInfo();
      }
      const accountId = this.account.id + "";
      this.eventSource = new EventSource("/api/events?t=" + accountId)
      this.eventSource.onmessage = (event) => {
        if (event.data === "info") {
          this.loadInfo();
        }
        if (accountId === event.data) {
          this.submitForm(null, this.entryQty);
        }
      };
      this.eventSource.onerror = (event) => {
        switch (event.target.readyState) {
          case EventSource.CONNECTING:
            break;
          case EventSource.CLOSED:
            this.listenToEvents();
            break;
        }
      }
    },
    loadInfo() {
      this.loading++;
      axios.get(apiUrl + "info").then((r) => {
        this.loading--;
        this.info = r.data;
        if (this.timer) clearTimeout(this.timer);
        this.initTime = new Date().getTime()
        this.timer = setInterval(() => {
          this.currentTime = new Date().getTime();
        }, 1000)
      });
    }
  },
  created() {
    this.loadInfo();
  },
  data() {
    return {
      initTime: null,
      currentTime: null,
      timer: null,
      counter: 0,
      info: null,
      form: {
        address: null,
        username: null,
      },
      entryQty: 1,
      account: null,
      error: {},
      loading: 0,
      eventSource: null
    };
  },
};
</script>

<template>
  <div class="bg-white">
    <faq :info="info" v-if="info" />
    <monero-related />
    <contact />
    <entries :info="info" v-if="info" />
    <winner :info="info" v-if="info && info.last_winner" />

    <div id="loader" v-if="loading > 0">
      <div class="bar"></div>
      <div class="bar"></div>
      <div class="bar"></div>
    </div>
    <div class="grid grid-cols-12 gap-0">
      <div
        class="bg-fixed relative col-span-12 sm:col-span-12 md:col-span-7 lg:col-span-8 xxl:col-span-8 hidden md:block"
      >
        <img src="./assets/bg.png" class="w-full" />
        <div
          class="absolute inset-0 z-20 flex items-center justify-center h-full bg-gray-900 bg-opacity-50"
        >
          <div class="flex text-center h-full items-center fixed top-0" v-if="info">
            <div class="border-2 p-4 rounded-md bg-gray-100 text-gray-500">
              Price Reset in {{ timers.until_price }}
              <div class="text-black text-2xl lg:text-5xl">{{ info.entry_price }}</div>
              {{ toUsd(info.entry_price) }}
            </div>
          </div>
        </div>
      </div>
      <div
        class="col-span-12 sm:col-span-12 md:col-span-5 lg:col-span-4 xxl:col-span-4"
        v-if="info"
      >
        <div>
          <div class="border-b">
            <div class="my-4 px-6">
              <h2 class="font-semibold text-2xl">
                Monero Pot
                <span
                  v-if="info.wallet_offline"
                  class="bg-red-600 p-2 rounded-lg text-white font-bold"
                >Wallet Offline</span>
              </h2>
            </div>
          </div>
          <div class="px-8 py-2">
            <h4 class="text-lg text-gray-500 font-thin">Next Draw</h4>
            <h4 class="font-semibold">{{ timers.until_draw }}</h4>
          </div>
          <div class="flex space-x-0 flex-col lg:flex-row lg:space-x-2 my-2 px-6">
            <div
              class="bg-gray-600 p-4 border-2 rounded-md shadow-lg w-full text-white text-center"
            >
              <h1 class="text-xl font-light">WIN XMR</h1>
              <h1 class="text-2xl text-green-100 font-semibold">
                {{ info.win_amount }}
                <div class="text-sm">{{ toUsd(info.win_amount) }}</div>
              </h1>
            </div>
            <div
              class="bg-yellow-700 p-4 border-2 rounded-md shadow-lg w-full text-white text-center"
            >
              <h1 class="text-xl font-light">REFERRAL XMR</h1>
              <h1 class="text-2xl text-red-100 font-semibold">
                {{ info.ref_amount }}
                <div class="text-sm">{{ toUsd(info.ref_amount) }}</div>
              </h1>
            </div>
          </div>
          <div class="px-8 my-6" v-if="account">
            <div class="my-4 border-b w-full" ref="accountInfo">
              <h2 class="font-semibold text-lg">{{ account.username || "Anonymous" }}</h2>
            </div>
            <div class="text-xs">
              <p class="text-center">Scan qr code or copy the address below into your wallet app.</p>
              <div class="flex justify-center">
                <QRCodeVue3 :value="addressUri" :width="200" :height="200" />
              </div>
              <p class="text-center pb-1">Number of Entries</p>
              <div class="text-center mb-3">
                <button type="button" class="button" @click="entryQty--">-</button>
                <input type="number" class="px-4 py-2 w-24 bg-gray-100" v-model.number="entryQty" />
                <button type="button" class="button" @click="entryQty++">+</button>
              </div>
              <div class="flex">
                <input
                  type="text"
                  readonly
                  class="rounded-sm px-4 py-3 mt-1 focus:outline-none bg-gray-100 w-full"
                  :value="account.address"
                />
                <button
                  type="button"
                  class="rounded-sm block text-center text-white bg-gray-700 px-2 my-1"
                  @click="copyToClipboard(account.address)"
                >Copy</button>
              </div>
              <template v-if="refUrl">
                <div class="mt-2">Referral Link:</div>
                <div class="flex">
                  <input
                    type="text"
                    readonly
                    class="rounded-sm px-4 py-3 mt-1 focus:outline-none bg-gray-100 w-full"
                    :value="refUrl"
                  />
                  <button
                    type="button"
                    class="rounded-sm block text-center text-white bg-gray-700 px-2 my-1"
                    @click="copyToClipboard(refUrl)"
                  >Copy</button>
                </div>
              </template>
            </div>
            <line-info label="Entry Price" :success="true">
              {{ info.entry_price }}
              <small class="text-gray-600">({{ toUsd(info.entry_price) }})</small>
            </line-info>
            <line-info label="Number of Entries" :success="account.entries > 0">
              <span
                :class="[
                  account.entries > 0 ? 'underline cursor-pointer' : null,
                ]"
                @click="account.entries > 0 ? viewEntries(account.id) : null"
              >{{ account.entries }}</span>
            </line-info>
            <line-info label="Unused XMR" :success="account.xmr > 0">
              {{ account.xmr }}
              <small class="text-gray-600">({{ toUsd(account.xmr) }})</small>
            </line-info>
            <line-info label="Referrals" :success="account.referrals > 0">{{ account.referrals }}</line-info>
            <div class="text-center">
              <a
                class="cursor-pointer underline text-sm"
                @click="toggleModal('MoneroRelated')"
              >Monero Related Things</a>
              <span class="text-sm pl-1">(to visit while you're waiting)</span>
            </div>
          </div>
          <div class="px-8 my-6">
            <div class="my-4 border-b w-full">
              <h2 class="font-semibold text-lg">Enter Your Monero Address</h2>
            </div>
            <div class="bg-white p-4 border-2 rounded-md">
              <form class="mt-4" @submit.prevent="submitForm">
                <div class="my-5 text-sm">
                  <label for="address" class="block text-black">
                    Your Monero Address*
                    <small class="text-gray-600" v-if="errorField('address')">
                      (
                      <span class="text-red-400">{{ errorField("address") }}</span>
                      )
                    </small>
                  </label>
                  <input
                    type="text"
                    id="address"
                    autofocus
                    class="rounded-sm px-4 py-3 mt-1 focus:outline-none bg-gray-100 w-full"
                    v-model="form.address"
                    placeholder="Monero Address"
                  />
                </div>
                <div class="my-5 text-sm">
                  <label for="username" class="block text-black">
                    Username
                    <small class="text-gray-600" v-if="errorField('username')">
                      (
                      <span class="text-red-400">{{ errorField("username") }}</span>
                      )
                    </small>
                  </label>
                  <input
                    id="username"
                    v-model="form.username"
                    type="text"
                    autofocus
                    class="rounded-sm px-4 py-3 mt-1 focus:outline-none bg-gray-100 w-full"
                    placeholder="Enter Username"
                  />
                  <small class="text-gray-500">Keep address for next drawing or for referrals</small>
                </div>
                <div class="my-5">
                  <button
                    :disabled="!formValid"
                    type="submit"
                    :class="{ 'opacity-50 cursor-not-allowed': !formValid }"
                    class="rounded-sm block text-center text-white bg-gray-800 p-3 duration-300 hover:bg-black w-full"
                  >View Entry Address &amp; Send Payment</button>
                </div>
              </form>
              <div class="text-center">
                <a class="cursor-pointer underline text-sm" @click="viewEntries(null)">Entries</a>
                &nbsp;
                <a
                  class="cursor-pointer underline text-sm"
                  @click="toggleModal('winner')"
                  v-if="info && info.last_winner"
                >Winner</a>
                &nbsp;
                <a
                  class="cursor-pointer underline text-sm"
                  @click="toggleModal('faq')"
                >FAQ</a>
                &nbsp;
                <a
                  class="cursor-pointer underline text-sm"
                  @click="toggleModal('contact')"
                >Contact</a>
                &nbsp;
                <a
                  target="_blank"
                  class="underline text-sm"
                  href="https://github.com/alagaris/moneropot"
                >Source</a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

