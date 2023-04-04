<template>
    <div class="home">
        <el-form ref="form" inline :model="form" :rules="rules" size="mini" label-width="120px" style="width: 100%">
            <el-form-item label="SimNo" prop="simNo">
                <el-autocomplete v-model="form.simNo" :fetch-suggestions="simNoSearch" placeholder="请输入SIM卡号"></el-autocomplete>
            </el-form-item>
            <el-form-item label="开始时间戳" prop="since">
                <el-date-picker v-model="form.since" type="datetime" placeholder="选择最早时间" align="right" :picker-options="pickerOptions"> </el-date-picker>
            </el-form-item>
            <el-form-item label="结束时间戳" prop="until">
                <el-date-picker v-model="form.until" type="datetime" placeholder="选择最晚时间" value-format="yyyy-MM-dd HH:mm:ss" align="right">
                </el-date-picker>
            </el-form-item>
            <el-form-item>
                <el-button type="primary" @click="onQuery">查询</el-button>
            </el-form-item>
        </el-form>
        <el-tabs tab-position="left" style="height: 100%; width: 100%">
            <el-tab-pane label="所有报文">
                <MsgView :data="data" />
            </el-tab-pane>
            <el-tab-pane label="定位报文">
                <MsgView0200 :data="data" />
            </el-tab-pane>
        </el-tabs>
    </div>
</template>

<script>
import moment from 'moment'
import MsgView from '@/components/MsgView.vue'
import MsgView0200 from '@/components/MsgView0200.vue'

const dateFormat = 'YYYY-MM-DD HH:mm:ss'

function createShortcut(name, seconds) {
    return {
        text: name,
        onClick(picker) {
            const time = new Date()
            time.setTime(time.getTime() - seconds * 1000)
            picker.$emit('pick', time)
        },
    }
}

export default {
    name: 'HomeView',
    components: {
        MsgView,
        MsgView0200
    },
    data() {
        return {
            data: [],
            loading: false,
            form: {
                simNo: '',
                since: moment().subtract(20, 'm').toDate(),
                until: null,
            },
            pickerOptions: {
                shortcuts: [createShortcut('20分钟前', 60 * 20), createShortcut('1小时前', 3600), createShortcut('1天前', 3600 * 24)],
            },
            rules: {
                simNo: [{ type: 'string', required: true, message: '请输入SimNo' }],
                since: [{ type: 'date', required: true, message: '请选择开始时间' }],
            },
        }
    },
    methods: {
        simNoSearch(q, cb) {
            cb(this.$store.state.simNoHistory.filter((it) => it.indexOf(q) == 0).map((it) => ({ value: it })))
        },
        onQuery() {
            this.$refs.form.validate(async (valid) => {
                if (!valid) {
                    return
                }
                this.$store.commit('addSimNoHistory', this.form.simNo)
                this.loading = true
                try {
                    const results = await this.$http.get('/query', {
                        params: {
                            simNo: this.form.simNo,
                            since: moment(this.form.since).format(dateFormat),
                            until: moment(this.form.until || new Date()).format(dateFormat),
                        },
                    })
                    this.data = results
                } finally {
                    this.loading = false
                }
            })
        },
    },
}
</script>

<style scoped lang="scss">
.home {
    background-color: #f0f0f0;
    padding: 20px;
    box-sizing: border-box;
    height: 100vh;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
}
.el-form {
    padding: 8px 0;
    background-color: #fff;
    border: 1px solid #ddd;
    box-sizing: border-box;
}
.el-tabs {
    width: 100%;
    border: 1px solid #ddd;
    box-sizing: border-box;
    background-color: #fff;
    margin-top: 4px;

    ::v-deep .el-tabs__content,
    ::v-deep .el-tab-pane {
        height: 100%;
    }
}
</style>
