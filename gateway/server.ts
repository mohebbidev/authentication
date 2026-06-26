import { buildApp } from './app'
import { config } from './config'

const app = buildApp()

const server = app.listen(config.port, () => {
    console.log(`Gateway listening on port ${config.port}`)
})

// Graceful shutdown
const shutdown = () => {
    console.log('Shutting down gateway...')
    server.close(() => {
        console.log('Gateway stopped')
        process.exit(0)
    })
}

process.on('SIGTERM', shutdown)
process.on('SIGINT', shutdown)