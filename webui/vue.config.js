const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  publicPath: '/ui/',
  transpileDependencies: true,
  productionSourceMap: false,
  devServer: {
    proxy: {
      "/api": {
        target: "http://192.168.73.30:6060"
      }
    }
  }
})
