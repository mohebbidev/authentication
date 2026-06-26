const required = (key: string): string => {
    const val = process.env[key]
    if (!val) throw new Error(`Missing required env var: ${key}`)
    return val
}

export const config = {
    port: process.env.PORT ?? '3000',

    grpc: {
        address: process.env.GRPC_ADDRESS ?? 'localhost:50051',
    },

    redis: {
        url: process.env.REDIS_URL ?? 'redis://localhost:6379',
    },

    jwt: {
        secret: required('JWT_SECRET'),
    },

    rateLimit: {
        // per IP, per window
        windowMs: 15 * 60 * 1000, // 15 minutes
        maxRequests: 100,
        // stricter limit for auth endpoints
        authWindowMs: 15 * 60 * 1000,
        authMaxRequests: 10,
    },

    cookie: {
        // how long the session cookie lives in the browser
        maxAgeMs: 30 * 24 * 60 * 60 * 1000, // 30 days
        secure: process.env.NODE_ENV === 'production',
        domain: process.env.COOKIE_DOMAIN ?? undefined,
    },
} as const