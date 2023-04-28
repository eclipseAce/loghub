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
            <el-form-item label="数据传输" prop="msgXfer">
                <el-select v-model="query.msgXfer" multiple placeholder="全部">
                    <el-option value="tx" label="下行" />
                    <el-option value="rx" label="上行" />
                </el-select>
            </el-form-item>
            <el-form-item label="报文类型" prop="msgIds">
                <el-select v-model="query.msgIds" multiple placeholder="全部">
                    <el-option v-for="item in msgIds" :key="item" :value="item" :label="item.toString(16).padStart(4, 0)" />
                </el-select>
            </el-form-item>
            <el-form-item>
                <el-button type="primary" :loading="loading" @click="onQuery">查询</el-button>
            </el-form-item>
        </el-form>
        <div class="view-wrapper" v-loading="loading">
            <DataTable :data="msgs" :filename="filename" style="height: 100%">
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
                <el-table-column prop="xferSign" label="" width="32" align="center">
                    <template slot-scope="{ row: { xferSign } }">
                        <i v-if="xferSign == '>>'" class="el-icon-d-arrow-right" style="color: #67c23a"></i>
                        <i v-if="xferSign == '<<'" class="el-icon-d-arrow-left" style="color: #e6a23c"></i>
                    </template>
                </el-table-column>
                <el-table-column prop="msgIdHex" label="ID" width="60" align="right"></el-table-column>
                <el-table-column prop="msgSn" label="SN" width="60" align="right"></el-table-column>
                <el-table-column prop="version" label="版本" width="60" align="right"></el-table-column>
                <el-table-column prop="part" label="分包" width="60" align="right"></el-table-column>
                <el-table-column prop="raw" label="数据"></el-table-column>
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
    name: 'QueryView',
    components: {
        SimNoInput,
        DataTable,
    },
    data() {
        return {
            msgs: [],
            msgIds: [],
            loading: false,
            filename: '',
            query: {
                simNo: '',
                since: moment().subtract(20, 'm').toDate(),
                until: null,
                msgIds: [],
                msgXfer: [],
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
    methods: {
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
                        msgIds: this.query.msgIds.join(','),
                        msgXfer: this.query.msgXfer.join(','),
                    }
                    const result = await this.$http.get('/query', { params })
                    this.msgs = result.msgs.reverse().map((it) =>
                        Object.assign(it, {
                            timestamp: moment(it.timestamp).format('YYYY-MM-DD HH:mm:ss'),
                            msgIdHex: it.msgId.toString(16).padStart(4, 0),
                            xferSign: it.tx ? '<<' : '>>',
                            version: it.version == -1 ? '2011' : `2019(v${it.version})`,
                            part: `${it.partIndex}/${it.partTotal}`,
                            raw: base64ToHex(it.raw),
                        })
                    )
                    this.msgIds = result.msgIds.sort((a, b) => a - b)
                    this.filename = `${params.simNo}_raw_${moment(params.since).format(filenameDateFormat)}-${moment(params.until).format(filenameDateFormat)}`
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
