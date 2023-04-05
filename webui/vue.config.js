const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  publicPath: '/ui/',
  transpileDependencies: true,
  devServer: {
    proxy: {
      "/api": {
        target: "http://192.168.73.30:6060"
      }
    }
  }
})
