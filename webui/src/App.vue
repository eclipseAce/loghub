<template>
    <div id="app">
        <el-menu :default-active="$route.path" mode="horizontal" router active-text-color="#409EFF">
            <el-menu-item index="/query">基本查询</el-menu-item>
            <el-menu-item index="/query0200">定位查询</el-menu-item>
            <el-menu-item index="/query0705">CAN查询</el-menu-item>
        </el-menu>
        <div style="width: 100%; height: calc(100vh - 61px)">
            <keep-alive>
                <router-view />
            </keep-alive>
        </div>
    </div>
</template>

<style lang="scss">
html {
    font-family: Avenir, Helvetica, Arial, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    height: 100%;
    overflow: hidden;
}
body,
#app {
    height: 100%;
    padding: 0;
    margin: 0;
}
</style>
