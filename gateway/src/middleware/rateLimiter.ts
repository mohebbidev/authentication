import { rateLimit } from 'express-rate-limit'
import { RedisStore } from 'rate-limit-redis'
import Redis from 'ioredis'
import { config } from '../../config'

let redisClient: Redis | null = null

function getRedis(): Redis {
    if (!redisClient) {
        redisClient = new Redis(config.redis.url)
        redisClient.on('error', (err) => {
            console.error('Redis error:', err)
        })
    }
    return redisClient
}

// General limiter — applied to all routes
export const generalLimiter = rateLimit({
    windowMs: config.rateLimit.windowMs,
    max: config.rateLimit.maxRequests,
    standardHeaders: 'draft-7',
    legacyHeaders: false,
    store: new RedisStore({
        // sendCommand: (...args: string[]) => getRedis().call(...args) as any,
        sendCommand: (...args: string[]) =>
            getRedis().call(...(args as [string, ...string[]])) as any,
        prefix: 'rl:general:',
    }),
    handler: (_req, res) => {
        res.status(429).json({ error: 'Too many requests, please try again later.' })
    },
})

// Strict limiter — applied to login, register, password reset
// 10 attempts per 15 minutes per IP
export const authLimiter = rateLimit({
    windowMs: config.rateLimit.authWindowMs,
    max: config.rateLimit.authMaxRequests,
    standardHeaders: 'draft-7',
    legacyHeaders: false,
    store: new RedisStore({
        // sendCommand: (...args: string[]) => getRedis().call(...args) as any,
        sendCommand: (...args: string[]) =>
            getRedis().call(...(args as [string, ...string[]])) as any,
        prefix: 'rl:auth:',
    }),
    handler: (_req, res) => {
        res.status(429).json({ error: 'Too many attempts, please try again in 15 minutes.' })
    },
})