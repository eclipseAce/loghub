const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  publicPath: '/ui/',
  transpileDependencies: true,
  devServer: {
    proxy: {
      "/api": {
        target: "http://localhost:6060"
      }
    }
  }
})
