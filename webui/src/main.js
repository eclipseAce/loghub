import Vue from 'vue'
import App from './App.vue'
import router from './router'

import ElementUI from 'element-ui'
import 'element-ui/lib/theme-chalk/index.css'

import Axios from 'axios'

Vue.config.productionTip = false

Vue.use(ElementUI)

Vue.prototype.$http = Axios.create({
    baseURL: '/api',
    timeout: 30000 * 1,
})
Vue.prototype.$http.interceptors.request.use(
    (config) => config,
    (error) => {
        ElementUI.Message({ message: error, type: 'error', offset: 80 })
        return Promise.reject(error)
    }
)
Vue.prototype.$http.interceptors.response.use(
    ({ data }) => {
        if (data.error) {
            ElementUI.Message({ message: data.error, type: 'error', offset: 80 })
            return Promise.reject(data)
        }
        return data.result
    },
    (error) => {
        ElementUI.Message({ message: error, type: 'error', offset: 80 })
        return Promise.reject(error)
    }
)

new Vue({
    router,
    render: (h) => h(App),
}).$mount('#app')
