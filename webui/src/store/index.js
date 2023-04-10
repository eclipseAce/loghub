import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

const store = new Vuex.Store({
    state: {
        simNoHistory: [],
    },
    getters: {},
    mutations: {
        initialiseStore(state) {
            // Check if the ID exists
            if (localStorage.getItem('store')) {
                // Replace the state object with the stored item
                this.replaceState(Object.assign(state, JSON.parse(localStorage.getItem('store'))))
            }
        },
        addSimNoHistory(state, simNo) {
            simNo = simNo.trim()
            if (!simNo) {
                return
            }
            const index = state.simNoHistory.indexOf(simNo)
            if (index > -1) {
                state.simNoHistory.splice(index, 1)
            }
            state.simNoHistory.push(simNo)
        },
    },
    actions: {},
    modules: {},
})

store.subscribe((mutation, state) => {
    // Store the state object as a JSON string
    localStorage.setItem('store', JSON.stringify(state))
})

export default store
