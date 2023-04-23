<template>
    <div class="view-wrapper">
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
            <el-table-column prop="Warnings" label="" width="32" align="center">
                <template slot-scope="{ row: { Warnings } }">
                    <el-tooltip v-if="Warnings.length !== 0" effect="dark" placement="right">
                        <i class="el-icon-warning table-row-icon" style="color: #f56c6c"></i>
                        <template slot="content">
                            <div v-for="(warning, i) in Warnings" :key="i" style="font-size: 14px">{{ warning }}</div>
                        </template>
                    </el-tooltip>
                </template>
            </el-table-column>
            <el-table-column prop="Xfer" label="" width="32" align="center">
                <template slot-scope="{ row: { TX } }">
                    <i v-if="!TX" class="el-icon-d-arrow-right table-row-icon" style="color: #67c23a"></i>
                    <i v-else class="el-icon-d-arrow-left table-row-icon" style="color: #e6a23c"></i>
                </template>
            </el-table-column>
            <el-table-column prop="Timestamp" label="时间戳" width="160" align="center"></el-table-column>
            <el-table-column prop="SimNo" label="SIM卡号" width="100" align="right"></el-table-column>
            <el-table-column prop="MsgID" label="ID" width="60" align="right"></el-table-column>
            <el-table-column prop="MsgSN" label="SN" width="60" align="right"></el-table-column>
            <el-table-column prop="Version" label="版本" width="60" align="right"></el-table-column>
            <el-table-column prop="Part" label="分包" width="60" align="right"></el-table-column>

            <template v-if="viewMode == '0200'">
                <el-table-column prop="Body.Time" label="定位时间" width="160" align="center"></el-table-column>
                <el-table-column prop="Body.Alarm" label="报警位" width="200" align="right"></el-table-column>
                <el-table-column prop="Body.Status" label="状态位" width="200" align="right"></el-table-column>
                <el-table-column prop="Body.Lnglat" label="经纬度" width="200" align="right"></el-table-column>
                <el-table-column prop="Body.Altitude" label="海拔(m)" width="100" align="right"></el-table-column>
                <el-table-column prop="Body.Speed" label="速度(km/h)" width="100" align="right"></el-table-column>
                <el-table-column prop="Body.Direction" label="航向(°)" width="100" align="right"></el-table-column>
            </template>

            <el-table-column v-if="viewMode == 'ALL'" prop="Raw" label="数据"></el-table-column>
        </el-table>
    </div>
</template>

<script>
import moment from 'moment'

const dateFormat = 'YYYY-MM-DD HH:mm:ss'

function base64ToHex(str) {
    const raw = atob(str)
    let result = []
    for (let i = 0; i < raw.length; i++) {
        const hex = raw.charCodeAt(i).toString(16)
        result.push(hex.length === 2 ? hex : '0' + hex)
    }
    return result.join(' ').toUpperCase()
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
    name: 'MsgView',
    props: {
        data: Array,
        simNo: String,
        since: Date,
        until: Date
    },
    data() {
        return {
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
                        MsgID: it.MsgID.toString(16).padStart(4, 0),
                    })
                })
                .reverse()
        },
        filterItems() {
            return this.items.filter((it) => {
                if ((it.TX && !this.msgXfer.tx) || (!it.TX && !this.msgXfer.rx)) {
                    return false
                }
                return this.visibleMsgIds.indexOf(it.MsgID) != -1
            })
        },
        pageItems() {
            return this.filterItems.slice((this.page.current - 1) * this.page.size, this.page.current * this.page.size).map((it) => {
                const item = Object.assign({}, it, {
                    Timestamp: moment(it.Timestamp).format(dateFormat),
                    Version: it.Version == -1 ? '-' : it.Version,
                    Part: `${it.PartIndex}/${it.PartTotal}`,
                    Raw: base64ToHex(it.Raw),
                    Body: {},
                })
                if (item.MsgID == '0200' && typeof it.Body === 'object') {
                    item.Body = Object.assign({}, it.Body, {
                        Alarm: formatBits(it.Body.Alarm),
                        Status: formatBits(it.Body.Status),
                        Lnglat: `${it.Body.Longitude.toFixed(6)},${it.Body.Latitude.toFixed(6)}`,
                        Speed: `${it.Body.Speed.toFixed(1)}`,
                        Time: moment(it.Body.Time).format(dateFormat),
                    })
                    item.Warnings = item.Warnings.concat(item.Body.Warnings)
                }
                return item
            })
        },
        visibleMsgIds() {
            return this.msgIds.filter((it) => it.checked).map((it) => it.value)
        },
        allMsgIdsChecked() {
            return this.msgIds.every((it) => it.checked)
        },
        viewMode() {
            if (this.visibleMsgIds.length == 1 && this.visibleMsgIds[0] == '0200') {
                return '0200'
            }
            return 'ALL'
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
                let msgId = msgIdMap[it.MsgID]
                if (!msgId) {
                    msgId = { value: it.MsgID, checked: true }
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
            const fileDateFormat = "YYYYMMDDHHmmss"
            saveData(
                this.filterItems.map(it => `${moment(it.Timestamp).format(dateFormat)}\t${base64ToHex(it.Raw)}`).join('\r\n'),
                `${this.simNo}_${moment(this.since).format(fileDateFormat)}-${moment(this.until).format(fileDateFormat)}.txt`,
                "text/plain"
            )
        },
    },
}
</script>

<style scoped lang="scss">
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
