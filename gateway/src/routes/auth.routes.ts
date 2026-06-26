import { Router, Request, Response } from 'express'
import { z } from 'zod'
import { AuthClient } from '../grpc/client'
import { validate } from '../middleware/validate'
import { requireAuth } from '../middleware/auth'
import { authLimiter } from '../middleware/rateLimiter'
import { handleGrpcError } from '../middleware/grpcError'
import { config } from '../../config'

const SESSION_COOKIE = 'sid'

const registerSchema = z.object({
    full_name: z.string().min(2).max(100),
    email: z.string().email(),
    password: z.string().min(8),
    date_of_birth: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Must be YYYY-MM-DD'),
})

const loginSchema = z.object({
    email: z.string().email(),
    password: z.string().min(1),
})

const verifyEmailSchema = z.object({
    token: z.string().min(1),
})

const requestResetSchema = z.object({
    email: z.string().email(),
})

const resetPasswordSchema = z.object({
    token: z.string().min(1),
    new_password: z.string().min(8),
})

export function authRouter(grpc: AuthClient): Router {
    const router = Router()

    // POST /auth/register
    router.post('/register', authLimiter, validate(registerSchema), async (req: Request, res: Response) => {
        try {
            const out = await grpc.register(req.body)
            res.status(201).json({ user_id: out.user_id })
        } catch (err) {
            handleGrpcError(err, res)
        }
    })

    // POST /auth/login
    // Sets httpOnly session cookie, returns access token in body.
    router.post('/login', authLimiter, validate(loginSchema), async (req: Request, res: Response) => {
        try {
            const out = await grpc.login(req.body)

            res.cookie(SESSION_COOKIE, out.session_id, {
                httpOnly: true,
                secure: config.cookie.secure,
                sameSite: 'strict',
                maxAge: config.cookie.maxAgeMs,
                domain: config.cookie.domain,
                path: '/',
            })

            res.json({ access_token: out.access_token })
        } catch (err) {
            handleGrpcError(err, res)
        }
    })

    // POST /auth/refresh
    // Client calls this when access token expires. Session cookie sent automatically.
    router.post('/refresh', async (req: Request, res: Response) => {
        const sessionId = req.cookies?.[SESSION_COOKIE]

        if (!sessionId) {
            res.status(401).json({ error: 'No active session' })
            return
        }

        try {
            const out = await grpc.refreshToken({ session_id: sessionId })
            res.json({ access_token: out.access_token })
        } catch (err) {
            // If session is expired/revoked, clear the stale cookie
            res.clearCookie(SESSION_COOKIE, { path: '/' })
            handleGrpcError(err, res)
        }
    })

    // POST /auth/logout
    // Requires valid session cookie — revokes it in Go, clears cookie here.
    router.post('/logout', requireAuth, async (req: Request, res: Response) => {
        const sessionId = req.cookies?.[SESSION_COOKIE]

        if (sessionId) {
            try {
                await grpc.logout({ session_id: sessionId })
            } catch {
                // Best effort — clear the cookie regardless
            }
        }

        res.clearCookie(SESSION_COOKIE, { path: '/' })
        res.json({ message: 'Logged out' })
    })

    // POST /auth/verify-email
    router.post('/verify-email', validate(verifyEmailSchema), async (req: Request, res: Response) => {
        try {
            await grpc.verifyEmail({ token: req.body.token })
            res.json({ message: 'Email verified successfully' })
        } catch (err) {
            handleGrpcError(err, res)
        }
    })

    // POST /auth/password-reset/request
    router.post('/password-reset/request', authLimiter, validate(requestResetSchema), async (req: Request, res: Response) => {
        try {
            await grpc.requestPasswordReset({ email: req.body.email })
            // Always return 200 — never reveal whether the email exists
            res.json({ message: 'If that email exists, a reset link has been sent.' })
        } catch {
            res.json({ message: 'If that email exists, a reset link has been sent.' })
        }
    })

    // POST /auth/password-reset/confirm
    router.post('/password-reset/confirm', authLimiter, validate(resetPasswordSchema), async (req: Request, res: Response) => {
        try {
            await grpc.resetPassword({
                token: req.body.token,
                new_password: req.body.new_password,
            })
            // Clear session cookie — user must log in again with new password
            res.clearCookie(SESSION_COOKIE, { path: '/' })
            res.json({ message: 'Password reset successfully. Please log in again.' })
        } catch (err) {
            handleGrpcError(err, res)
        }
    })

    return router
}