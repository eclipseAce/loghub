<template>
    <div class="view-wrapper">
        <div class="view-options">
            <span class="view-option-label">消息ID</span>
            <el-checkbox v-for="msgId in msgIds" :key="msgId.value" v-model="msgId.checked"
                :label="msgId.value"></el-checkbox>
        </div>
        <el-table :data="visibleItems" height="100%" stripe size="mini">
            <el-table-column prop="Warnings" label="" width="32" align="center">
                <template slot-scope="{ row: { Warnings } }">
                    <el-tooltip v-if="Warnings.length !== 0" effect="dark" placement="right">
                        <i class="el-icon-warning warning-icon"></i>
                        <template slot="content">
                            <div v-for="(warning, i) in Warnings" :key="i" style="font-size: 14px">{{ warning }}</div>
                        </template>
                    </el-tooltip>
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
            bits.push(i + 1)
        }
    }
    return bits.length === 0 ? '-' : bits.join(',')
}

export default {
    name: 'MsgView',
    props: {
        data: Array,
    },
    data() {
        return {
            msgIds: [],
        }
    },
    computed: {
        items() {
            return this.data
                .map((it) => {
                    const item = Object.assign({}, it, {
                        Timestamp: moment(it.Timestamp).format(dateFormat),
                        MsgID: it.MsgID.toString(16).padStart(4, 0),
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
                .reverse()
        },
        visibleItems() {
            return this.items.filter((it) => {
                return this.visibleMsgIds.indexOf(it.MsgID) != -1
            })
        },
        visibleMsgIds() {
            return this.msgIds.filter((it) => it.checked).map((it) => it.value)
        },
        viewMode() {
            if (this.visibleMsgIds.length == 1 && this.visibleMsgIds[0] == '0200') {
                return '0200'
            }
            return 'ALL'
        },
    },
    watch: {
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
    methods: {},
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

    &>.el-table {
        flex: 1 1;
    }

    ::v-deep .el-table__row {
        font-family: monospace;
        font-size: 13px;
        font-weight: 500;
        word-break: normal !important;
    }
    ::v-deep .el-table__cell {
        padding: 2px 0;
    }
}

.view-options {
    display: flex;
    align-items: center;
    box-sizing: border-box;
    width: 100%;
    height: 32px;
    padding: 0 8px;
    border-bottom: 1px solid #ddd;

    .view-option-label {
        font-size: 14px;
        color: #666;
        margin-right: 16px;
    }
}

.warning-icon {
    color: #f56c6c;
    font-size: 20px;
    vertical-align: middle;
    line-height: 20px;
}
</style>
