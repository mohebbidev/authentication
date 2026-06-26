import express from 'express'
import cookieParser from 'cookie-parser'
import helmet from 'helmet'
import pinoHttp from 'pino-http'
import { generalLimiter } from './middleware/rateLimiter'
import { authRouter } from './routes/auth'
import { getAuthClient } from './grpc/client'
import { config } from './config'

export function buildApp() {
    const app = express()

    // ── Security headers ────────────────────────────────────────────────────────
    app.use(helmet())

    // ── Structured request logging ──────────────────────────────────────────────
    app.use(pinoHttp({
        redact: ['req.headers.authorization', 'req.body.password', 'req.body.new_password'],
    }))

    // ── Body parsing ────────────────────────────────────────────────────────────
    app.use(express.json({ limit: '16kb' })) // small limit — auth payloads are tiny
    app.use(cookieParser())

    // ── Global rate limit ───────────────────────────────────────────────────────
    app.use(generalLimiter)

    // ── Routes ──────────────────────────────────────────────────────────────────
    const grpcClient = getAuthClient(config.grpc.address)
    app.use('/auth', authRouter(grpcClient))

    // ── Health check ────────────────────────────────────────────────────────────
    app.get('/health', (_req, res) => {
        res.json({ status: 'ok' })
    })

    // ── 404 handler ─────────────────────────────────────────────────────────────
    app.use((_req, res) => {
        res.status(404).json({ error: 'Not found' })
    })

    // ── Global error handler ─────────────────────────────────────────────────────
    app.use((err: Error, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
        console.error(err)
        res.status(500).json({ error: 'Internal server error' })
    })

    return app
}