import { Request, Response, NextFunction } from 'express'
import jwt from 'jsonwebtoken'
import { config } from '../config'

export interface AuthPayload {
    sub: string  // userID
    iat: number
    exp: number
}

// Extends Express Request so downstream handlers have req.user typed.
declare global {
    namespace Express {
        interface Request {
            user?: AuthPayload
        }
    }
}

// Verifies the JWT from the Authorization header.
// On success attaches the decoded payload to req.user and calls next().
// The Go service trusts the X-User-ID header the gateway sets — never trust raw JWT there.
export function requireAuth(req: Request, res: Response, next: NextFunction): void {
    const authHeader = req.headers.authorization

    if (!authHeader?.startsWith('Bearer ')) {
        res.status(401).json({ error: 'Missing or malformed Authorization header' })
        return
    }

    const token = authHeader.slice(7)

    try {
        const payload = jwt.verify(token, config.jwt.secret) as AuthPayload
        req.user = payload
        next()
    } catch (err) {
        if (err instanceof jwt.TokenExpiredError) {
            res.status(401).json({ error: 'Access token expired', code: 'TOKEN_EXPIRED' })
            return
        }
        res.status(401).json({ error: 'Invalid access token' })
    }
}