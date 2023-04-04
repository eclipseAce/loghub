const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  publicPath: '/ui/',
  transpileDependencies: true,
  devServer: {
    proxy: {
      "/api": {
        target: "http://127.0.0.1:6060"
      }
    }
  }
})
