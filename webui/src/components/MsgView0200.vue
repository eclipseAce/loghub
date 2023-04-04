<template>
    <div class="view-wrapper">
        <div class="view-options"></div>
        <el-table :data="items" height="100%" stripe size="mini">
            <el-table-column prop="timestamp" label="时间戳" width="160" align="center"></el-table-column>
            <el-table-column prop="simNo" label="SIM卡号" width="160" align="right"></el-table-column>

            <el-table-column prop="time" label="定位时间" width="160" align="center"></el-table-column>
            <el-table-column prop="alarm" label="报警位" width="200" align="right"></el-table-column>
            <el-table-column prop="status" label="状态位" width="200" align="right"></el-table-column>
            <el-table-column prop="lnglat" label="经纬度" width="200" align="right"></el-table-column>
            <el-table-column prop="altitude" label="海拔(m)" width="100" align="right"></el-table-column>
            <el-table-column prop="speed" label="速度(km/h)" width="100" align="right"></el-table-column>
            <el-table-column prop="direction" label="航向(°)" width="100" align="right"></el-table-column>

            <el-table-column v-if="columns.mileage" prop="mileage" label="里程(km)" width="100" align="right"></el-table-column>
            <el-table-column v-if="columns.fuel" prop="fuel" label="油量(L)" width="100" align="right"></el-table-column>
            <el-table-column v-if="columns.recorderSpeed" prop="recorderSpeed" label="记录仪速度(km/h)" width="120" align="right"></el-table-column>
            <el-table-column v-if="columns.signalStrength" prop="signalStrength" label="信号强度" width="80" align="right"></el-table-column>
            <el-table-column v-if="columns.satellites" prop="satellites" label="卫星数" width="80" align="right"></el-table-column>
            <el-table-column v-if="columns.analogValues" prop="analogValues" label="模拟值" width="80" align="right"></el-table-column>
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
            columns: {
                mileage: false,
                fuel: false,
                recorderSpeed: false,
                signalStrength: false,
                satellites: false,
                analogValues: false,
            },
        }
    },
    watch: {
        items(val) {
            const keys = Object.keys(this.columns)
            keys.forEach((k) => (this.columns[k] = false))
            val.forEach((it) => {
                keys.forEach((k) => {
                    if (it[k] !== null) {
                        this.columns[k] = true
                    }
                })
            })
        },
    },
    computed: {
        items() {
            return this.data
                .filter((it) => it.MsgID === 0x0200)
                .map((it) => {
                    const item = {
                        simNo: it.SimNo,
                        timestamp: moment(it.Timestamp).format(dateFormat),
                        warnings: it.Warnings,
                        raw: base64ToHex(it.Raw),
                    }
                    const body = it.DecodedBody
                    if (body !== null) {
                        const attach = body.AttachInfo
                        Object.assign(item, {
                            time: moment(body.Time).format(dateFormat),
                            alarm: formatBits(body.Alarm),
                            status: formatBits(body.Status),
                            lnglat: `${body.Longitude.toFixed(6)},${body.Latitude.toFixed(6)}`,
                            altitude: body.Altitude,
                            speed: `${body.Speed.toFixed(1)}`,
                            direction: `${body.Direction}`,

                            mileage: attach.Mileage !== null ? attach.Mileage.toFixed(1) : null,
                            fuel: attach.Fuel !== null ? attach.Fuel.toFixed(1) : null,
                            recorderSpeed: attach.RecorderSpeed !== null ? attach.RecorderSpeed.toFixed(1) : null,
                            signalStrength: attach.SignalStrength !== null ? attach.SignalStrength : null,
                            satellites: attach.Satellites !== null ? attach.Satellites : null,
                            analogValues: attach.AnalogValue0 !== null ? `${attach.AnalogValue0},${attach.AnalogValue1}` : null,
                        })
                    }
                    return item
                })
                .reverse()
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
        width: 100%;
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
