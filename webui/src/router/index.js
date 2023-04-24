import Vue from 'vue'
import VueRouter from 'vue-router'
import QueryView from '../views/QueryView.vue'

Vue.use(VueRouter)

const routes = [
    {
        path: '/',
        name: 'index',
        redirect: '/query',
    },
    {
        path: '/query',
        name: 'query',
        component: QueryView,
    },
]

const router = new VueRouter({
    routes,
})

export default router
