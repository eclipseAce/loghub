<template>
    <div class="query-view">
        <el-form class="query-form" ref="queryForm" :model="query" :rules="rules" size="mini" label-position="top">
            <el-form-item label="SimNo" prop="simNo">
                <SimNoInput ref="simNoInput" v-model="query.simNo" placeholder="请输入SIM卡号" />
            </el-form-item>
            <el-form-item label="开始时间戳" prop="since">
                <el-date-picker v-model="query.since" type="datetime" placeholder="选择最早时间" align="right" :picker-options="pickerOptions" />
            </el-form-item>
            <el-form-item label="结束时间戳" prop="until">
                <el-date-picker v-model="query.until" type="datetime" placeholder="选择最晚时间" align="right" />
            </el-form-item>
            <el-form-item>
                <el-button type="primary" :loading="loading" @click="onQuery">查询</el-button>
            </el-form-item>
        </el-form>

        <div class="view-wrapper" v-loading="loading">
            <DataTable :data="msgs" :filename="filename" style="height: 100%; flex: 1 1" :fit="false" highlight-current-row @current-change="onCurrentChange">
                <el-table-column prop="timestamp" label="时间戳" width="160" align="center"></el-table-column>
                <el-table-column prop="warnings" label="" width="32" align="center">
                    <template slot-scope="{ row: { warnings } }">
                        <el-tooltip v-if="warnings.length !== 0" effect="dark" placement="right">
                            <i class="el-icon-warning" style="color: #f56c6c"></i>
                            <template slot="content">
                                <div v-for="(warning, i) in warnings" :key="i" style="font-size: 14px">{{ warning }}</div>
                            </template>
                        </el-tooltip>
                    </template>
                </el-table-column>
                <el-table-column prop="time" label="接收时间" width="160" align="center"></el-table-column>
                <el-table-column prop="count" label="数量" width="100" align="right"></el-table-column>
                <el-table-column></el-table-column>
            </DataTable>

            <DataTable :data="canItems" :filename="canFilename" :fit="false" style="height: 100%; flex: 1 1; border-left: 1px solid #ddd">
                <el-table-column prop="id" label="CAN ID" width="100" align="center"></el-table-column>
                <el-table-column prop="data" label="CAN 数据" width="300" align="center"></el-table-column>
                <el-table-column></el-table-column>
            </DataTable>
        </div>
    </div>
</template>

<script>
import moment from 'moment'
import SimNoInput from '@/components/SimNoInput.vue'
import DataTable from '@/components/DataTable.vue'

const dateFormat = 'YYYY-MM-DD HH:mm:ss'
const filenameDateFormat = 'YYYYMMDDHHmmss'

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

function base64ToHex(str) {
    const raw = atob(str)
    let result = []
    for (let i = 0; i < raw.length; i++) {
        const hex = raw.charCodeAt(i).toString(16)
        result.push(hex.length === 2 ? hex : '0' + hex)
    }
    return result.join(' ').toUpperCase()
}

export default {
    name: 'Query0200View',
    components: {
        SimNoInput,
        DataTable,
    },
    data() {
        return {
            msgs: [],
            current: null,
            loading: false,
            filename: '',
            query: {
                simNo: '',
                since: moment().subtract(20, 'm').toDate(),
                until: null,
            },
            rules: {
                simNo: [{ type: 'string', required: true, message: '请输入SimNo' }],
                since: [{ type: 'date', required: true, message: '请选择开始时间' }],
            },
            pickerOptions: {
                shortcuts: [createShortcut('20分钟前', 60 * 20), createShortcut('1小时前', 3600), createShortcut('1天前', 3600 * 24)],
            },
        }
    },
    computed: {
        canItems() {
            return (this.current ? this.current.items : []).map((it) =>
                Object.assign({}, it, {
                    id: it.id.toString(16).padStart(8, 0).toUpperCase(),
                    data: base64ToHex(it.data),
                })
            )
        },
        canFilename() {
            if (!this.current) {
                return ''
            }
            return `${this.query.simNo}_0705@${moment(this.current.time).format('HHmmssSSS')}`
        },
    },
    methods: {
        onCurrentChange(current) {
            this.current = current
        },
        onQuery() {
            this.$refs.queryForm.validate(async (valid) => {
                if (!valid) {
                    return
                }
                this.$refs.simNoInput.appendHistory()
                this.loading = true
                try {
                    const params = {
                        simNo: this.query.simNo,
                        since: moment(this.query.since).format(dateFormat),
                        until: moment(this.query.until || new Date()).format(dateFormat), // default to now
                        msgId: 0x0705,
                    }
                    const result = await this.$http.get('/queryBody', { params })
                    this.msgs = result.reverse().map((it) =>
                        Object.assign(it, {
                            timestamp: moment(it.timestamp).format(dateFormat),
                            time: moment(it.time).format('HH:mm:ss.SSS'),
                        })
                    )
                    this.filename = `${params.simNo}_0705_${moment(params.since).format(filenameDateFormat)}-${moment(params.until).format(filenameDateFormat)}`
                } finally {
                    this.loading = false
                }
            })
        },
    },
}
</script>

<style scoped lang="scss">
.query-view {
    background-color: #f0f0f0;
    padding: 16px;
    box-sizing: border-box;
    height: 100%;
    display: flex;
}

.query-form {
    padding: 8px;
    background-color: #fff;
    border: 1px solid #ddd;
    box-sizing: border-box;
    width: 240px;
}
.query-form .el-form-item__content {
    & > .el-autocomplete,
    & > .el-select,
    & > .el-input,
    & > .el-date-editor,
    & > .el-button {
        width: 100%;
    }
}

.view-wrapper {
    display: flex;
    align-items: stretch;
    border: 1px solid #ddd;
    box-sizing: border-box;
    background-color: #fff;
    margin-left: 4px;
    width: calc(100% - 240px - 4px);
}
</style>
