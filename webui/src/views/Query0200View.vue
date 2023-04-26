<template>
    <div class="query-view">
        <el-form class="query-form" ref="queryForm" :model="query" :rules="rules" size="mini" label-position="top">
            <el-form-item label="SimNo" prop="simNo">
                <el-autocomplete v-model="query.simNo" :fetch-suggestions="simNoSearch" placeholder="请输入SIM卡号" />
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
            <el-table :data="pageMsgs" stripe size="mini" height="100%">
                <el-table-column type="index" width="60" :index="getPageItemIndex" align="right"></el-table-column>
                <el-table-column prop="timestamp" label="时间戳" width="160" align="center"></el-table-column>
                <el-table-column prop="warnings" label="" width="32" align="center">
                    <template slot-scope="{ row: { warnings } }">
                        <el-tooltip v-if="warnings.length !== 0" effect="dark" placement="right">
                            <i class="el-icon-warning table-row-icon" style="color: #f56c6c"></i>
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
            </el-table>

            <div class="view-options">
                <div style="flex: 1 1"></div>
                <el-pagination
                    class="view-options-item"
                    align="right"
                    @size-change="onPageSizeChange"
                    @current-change="onPageCurrentChange"
                    :current-page="page.current"
                    :page-size="page.size"
                    :page-sizes="[50, 100, 200, 500]"
                    layout="prev, pager, next, jumper, sizes, total"
                    :total="msgs.length"
                ></el-pagination>
                <div class="view-options-item">
                    <el-button type="primary" size="mini" :disabled="msgs.length == 0" @click="onDownload">下载TXT</el-button>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import moment from 'moment'

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

function formatBits(val) {
    const bits = []
    for (let i = 0; i < 32; i++) {
        if (val & (1 << i)) {
            bits.push(i)
        }
    }
    return bits.length === 0 ? '-' : bits.join(',')
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

const saveData = (function () {
    var a = document.createElement('a')
    document.body.appendChild(a)
    a.style = 'display: none'
    return function (content, fileName, type) {
        var blob = new Blob([content], { type: type }),
            url = window.URL.createObjectURL(blob)
        a.href = url
        a.download = fileName
        a.click()
        window.URL.revokeObjectURL(url)
    }
})()

export default {
    name: 'Query0200View',
    data() {
        return {
            msgs: [],
            loading: false,
            query: {
                simNo: '',
                since: moment().subtract(20, 'm').toDate(),
                until: null,
            },
            page: {
                current: 1,
                size: 100,
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
        pageMsgs() {
            return this.msgs.slice((this.page.current - 1) * this.page.size, this.page.current * this.page.size).map((it) => {
                const item = Object.assign({}, it, {
                    timestamp: moment(it.timestamp).format(dateFormat),
                    time: moment(it.time).format(dateFormat),
                    alarm: formatBits(it.alarm),
                    status: formatBits(it.status),
                    speed: it.speed.toFixed(1),
                    lnglat: `${it.longitude.toFixed(6)},${it.latitude.toFixed(6)}`
                })
                return item
            })
        },
    },
    methods: {
        onQuery() {
            this.$refs.queryForm.validate(async (valid) => {
                if (!valid) {
                    return
                }
                this.$store.commit('addSimNoHistory', this.query.simNo)
                this.loading = true
                try {
                    const params = {
                        simNo: this.query.simNo,
                        since: moment(this.query.since).format(dateFormat),
                        until: moment(this.query.until || new Date()).format(dateFormat), // default to now
                        msgId: 0x0200,
                    }
                    const result = await this.$http.get('/queryBody', { params })
                    this.msgs = result.reverse()
                    this.page.current = 1
                } finally {
                    this.loading = false
                }
            })
        },
        simNoSearch(q, cb) {
            cb(this.$store.state.simNoHistory.map((it) => ({ value: it })).reverse())
        },
        onPageSizeChange(val) {
            this.page.current = 1
            this.page.size = val
        },
        onPageCurrentChange(val) {
            this.page.current = val
        },
        getPageItemIndex(index) {
            return (this.page.current - 1) * this.page.size + index + 1
        },
        onDownload() {
            const fileDateFormat = 'YYYYMMDDHHmmss'
            saveData(
                this.msgs.map((it) => `${moment(it.timestamp).format(dateFormat)}\t${base64ToHex(it.raw)}`).join('\r\n'),
                `${this.query.simNo}_${moment(this.query.since).format(fileDateFormat)}-${moment(this.query.until).format(fileDateFormat)}.txt`,
                'text/plain'
            )
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

::v-deep .el-table__cell {
    padding: 2px 0;

    & > .cell {
        font-family: monospace;
        font-size: 13px;
        font-weight: 500;
        word-break: normal !important;
    }
}

.view-options {
    display: flex;
    align-items: stretch;
    box-sizing: border-box;
    width: 100%;
    padding: 8px 8px;
}
.view-options-item {
    display: flex;
    align-items: center;

    & + .view-options-item {
        margin-left: 16px;
        padding-left: 16px;
        border-left: 2px solid #ddd;
    }
}
.view-option-label {
    font-size: 14px;
    color: #666;
    margin-right: 16px;
    white-space: nowrap;
}
.view-option-content .el-checkbox {
    vertical-align: middle;
}
.table-row-icon {
    font-size: 20px;
    vertical-align: middle;
    line-height: 20px;
}
</style>
