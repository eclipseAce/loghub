<template>
    <div class="view-wrapper">
        <div class="view-options">
            <span class="view-option-label">消息ID</span>
            <el-checkbox v-for="msgId in msgIds" :key="msgId.value" v-model="msgId.checked" :label="msgId.value"></el-checkbox>
        </div>
        <el-table :data="filteredItems" height="100%" stripe size="mini">
            <el-table-column prop="timestamp" label="时间戳" width="160" align="center"></el-table-column>
            <el-table-column prop="simNo" label="SIM卡号" width="160" align="right"></el-table-column>
            <el-table-column prop="msgId" label="消息ID" width="80" align="center"></el-table-column>
            <el-table-column prop="msgSn" label="消息SN" width="80" align="right"></el-table-column>
            <el-table-column prop="version" label="消息版本" width="80" align="right"></el-table-column>
            <el-table-column prop="split" label="分包" width="80" align="right"></el-table-column>
            <el-table-column prop="raw" label="原始消息">
                <template slot-scope="{ row: { raw, warnings } }">
                    <div style="line-height: 16px">
                        <span style="vertical-align: middle">{{ raw }}</span>
                        <el-tooltip v-if="warnings.length !== 0" effect="dark" placement="left">
                            <i class="el-icon-warning" style="color: #f56c6c; font-size: 16px; vertical-align: middle; margin-right: 4px"></i>
                            <template slot="content">
                                <div v-for="(warning, i) in warnings" :key="i">{{ warning }}</div>
                            </template>
                        </el-tooltip>
                    </div>
                </template>
            </el-table-column>
        </el-table>
    </div>
</template>

<script>
import moment from 'moment'

const dateFormat = 'YYYY-MM-DD HH:mm:ss'

function base64ToHex(str) {
    const raw = atob(str)
    let result = ''
    for (let i = 0; i < raw.length; i++) {
        const hex = raw.charCodeAt(i).toString(16)
        result += hex.length === 2 ? hex : '0' + hex
    }
    return result.toUpperCase()
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
            return this.data.map((it) => ({
                simNo: it.SimNo,
                timestamp: moment(it.Timestamp).format(dateFormat),
                msgId: it.MsgID.toString(16).padStart(4, 0),
                msgSn: it.MsgSN,
                version: it.Version == -1 ? '-' : it.Version,
                split: `${it.PartIndex + 1}/${it.PartTotal}`,
                warnings: it.Warnings,
                raw: base64ToHex(it.Raw),
            })).reverse()
        },
        filteredItems() {
            const visibles = this.msgIds.filter((it) => it.checked).map((it) => it.value)
            return this.items.filter((it) => {
                return visibles.indexOf(it.msgId) != -1
            })
        },
    },
    watch: {
        items(value) {
            const map = {}
            this.msgIds = []
            value.forEach((it) => {
                if (map[it.msgId]) {
                    return
                }
                map[it.msgId] = true
                this.msgIds.push({ value: it.msgId, checked: true })
            })
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

    & > .el-table {
        flex: 1 1;
    }
    .el-table__row {
        font-family: monospace;
        font-size: 14px;
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
</style>
