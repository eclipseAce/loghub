const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  publicPath: '/ui/',
  transpileDependencies: true,
  devServer: {
    proxy: {
      "/api": {
        target: "http://192.168.70.129:6060"
      }
    }
  }
})
