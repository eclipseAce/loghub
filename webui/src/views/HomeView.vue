<template>
    <div class="home">
        <el-form ref="form" inline :model="form" :rules="rules" size="mini" label-width="120px">
            <el-form-item label="SimNo" prop="simNo">
                <el-input v-model="form.simNo"></el-input>
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
        <div class="view-wrapper">
            <div class="view-options">
                <span class="view-option-label">过滤消息ID</span>
                <el-checkbox v-for="msgId in msgIds" :key="msgId.value" v-model="msgId.checked" :label="msgId.value"></el-checkbox>
            </div>
            <el-table v-loading="loading" :data="filteredItems" height="100%" stripe size="mini" style="width: 100%; flex: 1 1">
                <el-table-column prop="timestamp" label="时间戳" width="160" align="center"></el-table-column>
                <el-table-column prop="simNo" label="SIM卡号" width="160" align="right"></el-table-column>
                <el-table-column prop="msgId" label="消息ID" width="80" align="center"></el-table-column>
                <el-table-column prop="msgSn" label="消息SN" width="80" align="right"></el-table-column>
                <el-table-column prop="version" label="消息版本" width="80" align="right">
                    <template slot-scope="{ row: { version } }">
                        <span v-if="version == -1">-</span>
                        <span v-else>{{ version }}</span>
                    </template>
                </el-table-column>
                <el-table-column prop="split" label="分包" width="80" align="right"></el-table-column>
                <el-table-column prop="raw" label="原始消息">
                    <template slot-scope="{ row: { raw, info } }">
                        <div style="line-height: 16px">
                            <span style="vertical-align: middle">{{ raw }}</span>
                            <el-tooltip v-if="info" effect="dark" :content="info" placement="left">
                                <i class="el-icon-warning" style="color: #f56c6c; font-size: 16px; vertical-align: middle; margin-right: 4px"></i>
                            </el-tooltip>
                        </div>
                    </template>
                </el-table-column>
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
    let result = ''
    for (let i = 0; i < raw.length; i++) {
        const hex = raw.charCodeAt(i).toString(16)
        result += hex.length === 2 ? hex : '0' + hex
    }
    return result.toUpperCase()
}

export default {
    name: 'HomeView',
    data() {
        return {
            items: [],
            msgIds: [],
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
    mounted() {},
    computed: {
        filteredItems() {
            const visibleMsgIds = this.msgIds.filter((it) => it.checked).map((it) => it.value)
            return this.items.filter((it) => {
                return visibleMsgIds.indexOf(it.msgId) != -1
            })
        },
    },
    methods: {
        onQuery() {
            this.$refs.form.validate(async (valid) => {
                if (!valid) {
                    return
                }
                this.loading = true
                try {
                    const results = await this.$http.get('/query', {
                        params: {
                            simNo: this.form.simNo,
                            since: moment(this.form.since).format(dateFormat),
                            until: moment(this.form.until || new Date()).format(dateFormat),
                        },
                    })
                    this.items = results
                        .map((it) => ({
                            simNo: it.SimNo,
                            timestamp: moment(it.Timestamp).format(dateFormat),
                            msgId: it.MsgID.toString(16).padStart(4, 0),
                            msgSn: it.MsgSN,
                            version: it.Version,
                            split: `${it.PartIndex + 1}/${it.PartTotal}`,
                            info: [`${it.BadChecksum ? '校验码错误' : ''}`, `${it.BadBodyLen ? '消息体长度错误' : ''}`].filter((inf) => inf !== '').join(';'),
                            raw: base64ToHex(it.Raw),
                        }))
                        .reverse()
                    const msgIdMap = {}
                    this.msgIds = []
                    this.items.forEach((it) => {
                        if (msgIdMap[it.msgId]) {
                            return
                        }
                        msgIdMap[it.msgId] = true
                        this.msgIds.push({ value: it.msgId, checked: true })
                    })
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
    border-radius: 4px;
    border: 1px solid #ddd;
    box-sizing: border-box;
}
.el-table {
    font-family: monospace;
    font-size: 14px;
}
.view-wrapper {
    flex: 1 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: flex-start;
    width: 100%;
    border-radius: 4px;
    border: 1px solid #ddd;
    box-sizing: border-box;
    background-color: #fff;

    & > .el-table {
        flex: 1 1;
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
