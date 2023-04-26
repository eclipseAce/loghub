import Vue from 'vue'
import VueRouter from 'vue-router'
import QueryView from '../views/QueryView.vue'
import Query0200View from '../views/Query0200View.vue'
import Query0705View from '../views/Query0705View.vue'

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
    {
        path: '/query0200',
        name: 'query0200',
        component: Query0200View,
    },
    {
        path: '/query0705',
        name: 'query0705',
        component: Query0705View,
    },
    { path: '*', redirect: '/' },
]

const router = new VueRouter({
    routes,
})

export default router
