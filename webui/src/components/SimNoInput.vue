<template>
    <el-autocomplete :value="value" v-on="$listeners" :fetch-suggestions="fetchHistory" />
</template>

<script>
export default {
    props: {
        value: String,
    },
    computed: {
        history() {
            return [...this.$store.state.simNoHistory].reverse()
        },
    },
    mounted() {
        if (this.history.length > 0) {
            this.$emit('input', this.history[0])
        }
    },
    methods: {
        fetchHistory(q, cb) {
            cb(this.history.map((it) => ({ value: it })))
        },
        appendHistory() {
            this.$store.commit('addSimNoHistory', this.value)
        },
    },
}
</script>
