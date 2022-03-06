import { createApp } from 'vue'
import App from './App.vue'
import './assets/tailwind.css'
import mitt from 'mitt';
const eventBus = mitt();
const app = createApp(App);
app.config.globalProperties.eventBus = eventBus;
app.config.globalProperties.appUtil = {
    toUsd(rate, value) {
        return (
            "$" +
            (Number(rate) * Number(value))
                .toFixed(2)
                .replace(/\B(?=(\d{3})+(?!\d))/g, ",")
        );
    },
};
app.mount('#app')
