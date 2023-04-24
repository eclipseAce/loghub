<template>
    <div class="query-view">
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
        <div class="view-wrapper" v-loading="loading">
            <div class="view-options">
                <div class="view-options-item" style="width: 240px">
                    <span class="view-option-label">消息传输</span>
                    <div class="view-option-content">
                        <el-checkbox v-model="msgXfer.rx" label="上行"></el-checkbox>
                        <el-checkbox v-model="msgXfer.tx" label="下行"></el-checkbox>
                    </div>
                </div>
                <div class="view-options-item" style="flex: 1 1">
                    <span class="view-option-label">消息ID</span>
                    <div class="view-option-content">
                        <el-checkbox :value="allMsgIdsChecked" @input="checkAllMsgIds" label="全部"></el-checkbox>
                        <el-checkbox v-for="msgId in msgIds" :key="msgId.value" v-model="msgId.checked" :label="msgId.value"></el-checkbox>
                    </div>
                </div>
                <el-pagination
                    class="view-options-item"
                    align="right"
                    @size-change="onPageSizeChange"
                    @current-change="onPageCurrentChange"
                    :current-page="page.current"
                    :page-size="page.size"
                    :page-sizes="[50, 100, 200, 500]"
                    layout="prev, pager, next, jumper, sizes, total"
                    :total="filterItems.length"
                ></el-pagination>
                <div class="view-options-item">
                    <el-button type="primary" size="mini" :disabled="data.length == 0" @click="onDownload">下载TXT</el-button>
                </div>
            </div>

            <el-table :data="pageItems" height="100%" stripe size="mini" style="flex: 1 1">
                <el-table-column type="index" width="60" :index="getPageItemIndex" align="right"></el-table-column>
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
                <el-table-column prop="xfer" label="" width="32" align="center">
                    <template slot-scope="{ row: { tx } }">
                        <i v-if="!tx" class="el-icon-d-arrow-right table-row-icon" style="color: #67c23a"></i>
                        <i v-else class="el-icon-d-arrow-left table-row-icon" style="color: #e6a23c"></i>
                    </template>
                </el-table-column>
                <el-table-column prop="timestamp" label="时间戳" width="160" align="center"></el-table-column>
                <el-table-column prop="simNo" label="SIM卡号" width="100" align="right"></el-table-column>
                <el-table-column prop="msgId" label="ID" width="60" align="right"></el-table-column>
                <el-table-column prop="msgSn" label="SN" width="60" align="right"></el-table-column>
                <el-table-column prop="version" label="版本" width="60" align="right"></el-table-column>
                <el-table-column prop="part" label="分包" width="60" align="right"></el-table-column>
                <el-table-column prop="raw" label="数据"></el-table-column>
            </el-table>
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
    name: 'HomeView',
    data() {
        return {
            data: [],
            query: {
                simNo: '',
                since: null,
                until: null,
            },
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
            msgIds: [],
            msgXfer: {
                rx: true,
                tx: true,
            },
            page: {
                current: 1,
                size: 100,
            },
        }
    },
    computed: {
        items() {
            return this.data
                .map((it) => {
                    return Object.assign({}, it, {
                        msgId: it.msgId.toString(16).padStart(4, 0),
                    })
                })
                .reverse()
        },
        filterItems() {
            return this.items.filter((it) => {
                if ((it.tx && !this.msgXfer.tx) || (!it.tx && !this.msgXfer.rx)) {
                    return false
                }
                return this.visibleMsgIds.indexOf(it.msgId) != -1
            })
        },
        pageItems() {
            return this.filterItems.slice((this.page.current - 1) * this.page.size, this.page.current * this.page.size).map((it) => {
                const item = Object.assign({}, it, {
                    timestamp: moment(it.timestamp).format(dateFormat),
                    version: it.version == -1 ? '-' : it.version,
                    part: `${it.partIndex}/${it.partTotal}`,
                    raw: base64ToHex(it.raw),
                })
                return item
            })
        },
        visibleMsgIds() {
            return this.msgIds.filter((it) => it.checked).map((it) => it.value)
        },
        allMsgIdsChecked() {
            return this.msgIds.every((it) => it.checked)
        },
    },
    watch: {
        data() {
            this.page.current = 1
        },
        items(value) {
            const msgIdMap = {}
            this.msgIds.forEach((it) => {
                msgIdMap[it.value] = it
            })
            const occurs = {}
            value.forEach((it) => {
                let msgId = msgIdMap[it.msgId]
                if (!msgId) {
                    msgId = { value: it.msgId, checked: true }
                    msgIdMap[msgId.value] = msgId
                }
                occurs[msgId.value] = true
            })
            this.msgIds = Object.values(msgIdMap)
                .filter((it) => !!occurs[it.value])
                .sort((a, b) => a.value.localeCompare(b.value))
        },
    },
    methods: {
        simNoSearch(q, cb) {
            cb(this.$store.state.simNoHistory.map((it) => ({ value: it })).reverse())
        },
        onQuery() {
            this.$refs.form.validate(async (valid) => {
                if (!valid) {
                    return
                }
                this.$store.commit('addSimNoHistory', this.form.simNo)
                this.loading = true
                try {
                    Object.assign(this.query, {
                        simNo: this.form.simNo,
                        since: this.form.since,
                        until: this.form.until || new Date(),
                    })
                    const results = await this.$http.get('/query', {
                        params: {
                            simNo: this.query.simNo,
                            since: moment(this.query.since).format(dateFormat),
                            until: moment(this.query.until).format(dateFormat),
                        },
                    })
                    this.data = results
                } finally {
                    this.loading = false
                }
            })
        },
        checkAllMsgIds(val) {
            this.msgIds.forEach((it) => (it.checked = val))
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
                this.filterItems.map((it) => `${moment(it.timestamp).format(dateFormat)}\t${base64ToHex(it.raw)}`).join('\r\n'),
                `${this.form.simNo}_${moment(this.form.since).format(fileDateFormat)}-${moment(this.form.until).format(fileDateFormat)}.txt`,
                'text/plain'
            )
        },
    },
}
</script>

<style scoped lang="scss">
.query-view {
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

.view-wrapper {
    flex: 1 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: flex-start;
    height: 100%;
    width: 100%;
    border: 1px solid #ddd;
    box-sizing: border-box;
    background-color: #fff;
    margin-top: 4px;
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
    border-bottom: 1px solid #ddd;
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
