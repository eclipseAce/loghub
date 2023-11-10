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
            <el-form-item label="包含附加信息" prop="extIds">
                <el-select v-model="query.extIds" multiple placeholder="任意">
                    <el-option v-for="item in extIds" :key="item" :value="item" :label="formatExtId(item)" />
                </el-select>
            </el-form-item>
            <el-form-item>
                <el-button type="primary" :loading="loading" @click="onQuery">查询</el-button>
            </el-form-item>
        </el-form>
        <div class="view-wrapper" v-loading="loading">
            <DataTable :data="msgs" :filename="filename" style="flex: 1 1">
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
                <el-table-column prop="time" label="定位时间" width="160" align="center"></el-table-column>
                <el-table-column prop="alarm" label="报警位" width="200" align="right"></el-table-column>
                <el-table-column prop="status" label="状态位" width="200" align="right"></el-table-column>
                <el-table-column prop="lnglat" label="经纬度" width="200" align="right"></el-table-column>
                <el-table-column prop="altitude" label="海拔(m)" width="100" align="right"></el-table-column>
                <el-table-column prop="speed" label="速度(km/h)" width="100" align="right"></el-table-column>
                <el-table-column prop="direction" label="航向(°)" width="100" align="right"></el-table-column>
                <el-table-column prop="mileage" label="里程(km)" width="100" align="right"></el-table-column>
                <el-table-column
                    v-for="extId in query.extIds"
                    :key="extId"
                    :prop="`extInfo_${formatExtId(extId)}`"
                    :label="`附加信息(${formatExtId(extId)})`"
                    width="200px"
                    align="left"
                ></el-table-column>
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

function formatBits(val) {
    const bits = []
    for (let i = 0; i < 32; i++) {
        if (val & (1 << i)) {
            bits.push(i)
        }
    }
    return bits.length === 0 ? '-' : bits.join(',')
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
            loading: false,
            query: {
                simNo: '',
                since: moment().subtract(20, 'm').toDate(),
                until: null,
                extIds: [],
            },
            rules: {
                simNo: [{ type: 'string', required: true, message: '请输入SimNo' }],
                since: [{ type: 'date', required: true, message: '请选择开始时间' }],
            },
            pickerOptions: {
                shortcuts: [createShortcut('20分钟前', 60 * 20), createShortcut('1小时前', 3600), createShortcut('1天前', 3600 * 24)],
            },
            filename: '',
            extIds: [],
        }
    },
    methods: {
        formatExtId(id) {
            return id.toString(16).padStart(2, 0).toUpperCase()
        },
        formatExtData(extInfo, id) {
            return extInfo
                .filter((it) => it.id === id)
                .map((it) => this.base64ToHex(it.data))
                .join('\r\n')
        },
        base64ToHex(str) {
            const raw = atob(str)
            let result = []
            for (let i = 0; i < raw.length; i++) {
                const hex = raw.charCodeAt(i).toString(16)
                result.push(hex.length === 2 ? hex : '0' + hex)
            }
            return result.join(' ').toUpperCase()
        },
        onQuery() {
            this.$refs.queryForm.validate(async (valid) => {
                if (!valid) {
                    return
                }
                this.$refs.simNoInput.appendHistory()
                this.loading = true
                this.extIds = []
                try {
                    const params = {
                        simNo: this.query.simNo,
                        since: moment(this.query.since).format(dateFormat),
                        until: moment(this.query.until || new Date()).format(dateFormat), // default to now
                        msgId: 0x0200,
                        extIds: this.query.extIds.join(',')
                    }
                    const result = await this.$http.get('/queryBody', { params })
                    this.msgs = result.reverse().map((it) => {
                        const row = Object.assign(it, {
                            timestamp: moment(it.timestamp).format(dateFormat),
                            time: moment(it.time).format(dateFormat),
                            alarm: formatBits(it.alarm),
                            status: formatBits(it.status),
                            speed: it.speed.toFixed(1),
                            lnglat: `${it.longitude.toFixed(6)},${it.latitude.toFixed(6)}`,
                        })
                        it.extInfo.forEach((ei) => (row[`extInfo_${this.formatExtId(ei.id)}`] = this.base64ToHex(ei.data)))
                        return row
                    })
                    const extIds = {}
                    result.forEach((it) => it.extInfo.forEach((extInfo) => (extIds[extInfo.id] = true)))
                    this.extIds = Object.keys(extIds).map((it) => parseInt(it))
                    this.filename = `${params.simNo}_0200_${moment(params.since).format(filenameDateFormat)}-${moment(params.until).format(filenameDateFormat)}`
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
    flex-direction: column;
    justify-content: center;
    align-items: stretch;
    border: 1px solid #ddd;
    box-sizing: border-box;
    background-color: #fff;
    margin-left: 4px;
    width: calc(100% - 240px - 4px);
}
</style>
