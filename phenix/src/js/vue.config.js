module.exports = {
  assetsDir: 'assets',

  devServer: {
    proxy: {
      '/api/v1': {
        target: 'http://darkbuntu:3001',
        changeOrigin: true,
        logLevel: 'debug',
        ws: true
      }
    }
  }
}
