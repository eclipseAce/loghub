<template>
    <div class="data-table">
        <el-table ref="table" stripe size="mini" v-bind="$attrs" v-on="$listeners" :data="page" height="calc(100% - 48px)">
            <el-table-column type="index" width="60" :index="getItemIndex" align="right"></el-table-column>
            <slot></slot>
        </el-table>
        <div class="data-table-footer">
            <div style="flex: 1 1"></div>
            <div class="data-table-footer-item">
                <el-pagination
                    align="right"
                    @size-change="onSizeChange"
                    @current-change="onCurrentChange"
                    :current-page="current"
                    :page-size="size"
                    :page-sizes="sizes"
                    :pager-count="5"
                    layout="prev, pager, next, total, sizes"
                    :total="data.length"
                ></el-pagination>
            </div>
            <div class="data-table-footer-item">
                <el-button type="primary" size="mini" :disabled="data.length == 0" @click="onDownload">下载CSV</el-button>
            </div>
        </div>
    </div>
</template>

<script>
import exportFromJSON from 'export-from-json'

export default {
    name: 'DataTable',
    props: {
        data: Array,
        filename: String,
    },
    data() {
        return {
            current: 1,
            size: 100,
            sizes: [50, 100, 200, 500],
        }
    },
    watch: {
        data() {
            this.current = 1
        },
    },
    computed: {
        page() {
            return this.data.slice((this.current - 1) * this.size, this.current * this.size)
        },
    },
    methods: {
        onSizeChange(val) {
            this.current = 1
            this.size = val
        },
        onCurrentChange(val) {
            this.current = val
        },
        getItemIndex(index) {
            return (this.current - 1) * this.size + index + 1
        },
        onDownload() {
            const fields = this.$refs.table.columns.map((c) => c.property).filter((p) => p)
            exportFromJSON({
                data: this.data,
                fields: fields,
                fileName: this.filename,
                exportType: 'csv',
                withBOM: true,
                beforeTableEncode(entries) {
                    return fields.map((field) => entries.find((e) => e.fieldName === field))
                },
            })
        },
    },
}
</script>

<style lang="scss" scoped>
::v-deep .el-table__cell {
    padding: 2px 0;

    & > .cell {
        font-family: monospace;
        font-size: 13px;
        font-weight: 500;
        word-break: normal !important;

        i[class^='el-icon-'] {
            font-size: 20px;
            vertical-align: middle;
            line-height: 20px;
        }
    }
}
.data-table-footer {
    display: flex;
    align-items: stretch;
    box-sizing: border-box;
    padding: 8px 8px;
    width: 100%;
    height: 48px;

    .data-table-footer-item {
        display: flex;
        align-content: center;

        & + .data-table-footer-item {
            margin-left: 16px;
            padding-left: 16px;
            border-left: 2px solid #ddd;
        }
    }
}
</style>
